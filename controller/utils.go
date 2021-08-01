package controller

import (
	"errors"
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"net"
	"strconv"
	"strings"
	"unicode"
)

func GetPodInfo() *SetRuleConfig{

	return nil
}

func SetTcRule(cfg *SetRuleConfig){
	// check network mode
	if cfg.HostNetwork {
		LimitPort(GetHostLink())
	}
	//set Ingress
	if len(cfg.Ingress) > 0 {
		if len(cfg.HostVETHName) == 0 {
			log.Fatalf("can not get host veth name!")
		}
		log.Printf("the host veth name is %s",cfg.HostVETHName)
		hostVethLink, err := netlink.LinkByName(cfg.HostVETHName)
		if err != nil {
			log.Fatalf("error found link %s,%w",cfg.HostVETHName,err)
		}
		_ , err = EnsureLinkUp(hostVethLink)
		if err != nil {
			log.Fatalf("error set link %s to up, %w",hostVethLink,err)
		}
		rate , err := Translate(cfg.Ingress)
		if err != nil {
			log.Fatalf("error get rate %s",err)
		}
		ReplaceTbf(hostVethLink,rate)
	}

	//set Egress

}

//EnsureLinkUp set link up, return changed and err
func EnsureLinkUp(link netlink.Link) (bool,error) {
	if link.Attrs().Flags&net.FlagUp != 0 {
		return false,nil
	}
	return true,LinkSetUp(link)
}

func LinkSetUp(link netlink.Link) error {
	errInfo := fmt.Sprintf("ip link set %s up",link.Attrs().Name)
	err := netlink.LinkSetUp(link)
	if err!=nil{
		return fmt.Errorf("error %s,%w",errInfo,err)
	}
	return nil
}


//translate ingress/egress to tc rate
func Translate(tf string) (uint64,error) {
	f := func(r rune) bool {
		return unicode.IsLetter(r)
	}
	index := strings.IndexFunc(tf,f)
	digit ,err := strconv.Atoi(tf[0:index])
	if err != nil{
		return 0, err
	}
	unit := tf[index:]
	unit = strings.ToLower(unit)
	var rate uint64
	switch unit {
	case "bit":
		rate = pow(digit,3)
	case "kbit":
		rate = pow(digit,6)
	case "mbit":
		rate = pow(digit,9)
	case "gbit":
		rate = pow(digit,12)
	case "tbit":
		rate = pow(digit,15)
	case "bps":
		rate = pow(digit,3)*8
	case "kbps":
		rate = pow(digit,6)*8
	case "mbps":
		rate = pow(digit,9)*8
	case "gbps":
		rate = pow(digit,12)*8
	case "tbps":
		rate = pow(digit,15)*8
	default:
		err = errors.New("invalid unit")
	}
	if err != nil {
		return 0,err
	}
	return rate,nil
}

func pow(data,n int) uint64 {
	for i:=0;i<n;i++ {
		data*=10
	}
	return uint64(data)
}