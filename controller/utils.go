package controller

import (
	"errors"
	"github.com/containernetworking/plugins/pkg/ns"
	jsoniter "github.com/json-iterator/go"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
	"unicode"
)


func SetTcRule(cfg *SetRuleConfig) {
	// check network mode
	if cfg.HostNetwork {
		LimitPort(GetHostLink())
	}

	//set Ingress on host veth
	if len(cfg.Ingress) > 0 {
		if cfg.HostVethIndex == -1 {
			klog.Errorf("can not get host veth !")
		}

		link, err := netlink.LinkByIndex(cfg.HostVethIndex)
		if err != nil {
			klog.Errorf("error found link %s,%+v", cfg.HostVethIndex, err)
		}

		hostVethLink := link.(*netlink.Veth)
		klog.Infof("the host veth name is %s", hostVethLink.Attrs().Name)

		_, err = EnsureLinkUp(hostVethLink)
		if err != nil {
			klog.Errorf("error set link %s to up, %+v", hostVethLink, err)
		}
		rate, err := Translate(cfg.Ingress)
		klog.Infof("the rate translated is %d bytes per second", rate)

		if err != nil {
			klog.Errorf("error get rate, error: %s", err)
		}

		err = ReplaceTbf(hostVethLink, rate)
		if err != nil {
			klog.Errorf("set qdisc on link %s failed!",hostVethLink.Name)
		}else {
			klog.Infof("set qdisc on host veth %d: %s successfully",hostVethLink.Index,hostVethLink.Name)
		}

	}

	//set Egress on container veth
	if len(cfg.Egress) > 0{
		if cfg.ContVethIndex == -1 {
			klog.Error("can not get container's veth!")
			return
		}
		if cfg.containerNetNs == nil {
			klog.Error("can not get container's net namespace!")
			return
		}

		cfg.containerNetNs.Do(func(_ ns.NetNS) error {
			link,err:= netlink.LinkByIndex(cfg.ContVethIndex)
			if err != nil {
				klog.Error("get container's link failed!")
				return err
			}
			contVeth := link.(*netlink.Veth)

			_, err = EnsureLinkUp(contVeth)
			if err != nil {
				klog.Errorf("error set link %s to up, %+v", contVeth, err)
			}
			rate, err := Translate(cfg.Egress)
			klog.Infof("the rate translated is %d bytes per second", rate)

			if err != nil {
				klog.Errorf("error get rate, error: %s", err)
			}

			err = ReplaceTbf(contVeth, rate)
			if err != nil {
				klog.Errorf("set qdisc on link %s failed!",contVeth.Name)
			}else {
				klog.Infof("set qdisc on container's veth %d: %s successfully",contVeth.Index,contVeth.Name)
			}

			return err
		})
	}

}

//translate ingress/egress to tc rate
func Translate(tf string) (uint64, error) {
	f := func(r rune) bool {
		return unicode.IsLetter(r)
	}
	index := strings.IndexFunc(tf, f)
	digit, err := strconv.Atoi(tf[0:index])
	if err != nil {
		return 0, err
	}
	unit := tf[index:]
	unit = strings.ToLower(unit)
	var rate uint64
	switch unit {
	case TC_BPS:
		break
	case TC_KPS:
		rate = Pow(digit, 3)
	case TC_MPS:
		rate = Pow(digit, 6)
	case TC_GPS:
		rate = Pow(digit, 9)
	case TC_TPS:
		rate = Pow(digit, 12)
	default:
		err = errors.New("invalid unit")
	}
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func Pow(data, n int) uint64 {
	for i := 0; i < n; i++ {
		data *= 10
	}
	return uint64(data)
}


//When a json object is large and has many fields. It is absolutely unnecessary to
//create a large entity class object in order to access several fields.
func GetFiledFromJason(path []string, data []byte) string {
	var temp jsoniter.Any
	for i,v := range path {
		if i==0 {
			temp = jsoniter.Get(data,v)
			if temp == nil {
				return ""
			}
		}else {
			temp = temp.Get(v)
			if temp == nil {
				return ""
			}
		}
	}

	switch temp.ValueType() {
	case jsoniter.InvalidValue,jsoniter.NilValue,jsoniter.BoolValue,jsoniter.ArrayValue,jsoniter.ObjectValue:
		return ""
	case jsoniter.StringValue:
		return temp.ToString()
	case jsoniter.NumberValue:
		return strconv.Itoa(temp.ToInt())
	}

	return ""
}


