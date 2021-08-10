package controller

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
	"math"
)

func burst(rate uint64, mtu int) uint32 {
	return uint32(math.Ceil(math.Max(float64(rate)/netlink.Hz(), float64(mtu))))
}

func time2Tick(time uint32) uint32 {
	return uint32(float64(time) * float64(netlink.TickInUsec()))
}

func buffer(rate uint64, burst uint32) uint32 {
	return time2Tick(uint32(float64(burst) * float64(netlink.TIME_UNITS_PER_SEC) / float64(rate)))
}

func limit(rate uint64, latency float64, buffer uint32) uint32 {
	return uint32(float64(rate)*latency/float64(netlink.TIME_UNITS_PER_SEC)) + buffer
}

func latencyInUsec(latencyInMillis float64) float64 {
	return float64(netlink.TIME_UNITS_PER_SEC) * (latencyInMillis / 1000.0)
}

const latencyInMillis = 25
const hardwareHeaderLen = 1500

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

//tc qdisc replace dev $dev root tbf rate $rate latency 50ms burst 20k
//unit is byte
func ReplaceTbf(dev netlink.Link, rate uint64) error {
	if rate <= 0 {
		return fmt.Errorf("invalid rate #{rate}")
	}
	burst := burst(rate, dev.Attrs().MTU+hardwareHeaderLen)
	buffer := buffer(rate, burst)
	latency := latencyInUsec(latencyInMillis)
	limit := limit(rate, latency, buffer)

	tbf := &netlink.Tbf{
		QdiscAttrs: netlink.QdiscAttrs{
			LinkIndex: dev.Attrs().Index,
			Handle:    netlink.MakeHandle(1, 0),
			Parent:    netlink.HANDLE_ROOT,
		},
		Rate:   rate,
		Limit:  uint32(limit),
		Buffer: uint32(buffer),
	}

	if err := netlink.QdiscReplace(tbf); err != nil {
		return errors.Wrapf(err, "can not replace qdics %+v on device %v/%s", tbf, dev.Attrs().Namespace, dev.Attrs().Name)
	}
	return nil
}

//tc qdisc del dev $dev root
func DeleteRule(qdisc netlink.Qdisc) error {
	err := netlink.QdiscDel(qdisc)
	return err
}
