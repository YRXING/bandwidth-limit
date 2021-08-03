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

func GetPodInfo() *SetRuleConfig {

	return nil
}

func SetTcRule(cfg *SetRuleConfig) {
	// check network mode
	if cfg.HostNetwork {
		LimitPort(GetHostLink())
	}
	//set Ingress
	if len(cfg.Ingress) > 0 {
		if len(cfg.HostVETHName) == 0 {
			log.Fatalf("can not get host veth name!")
		}
		log.Printf("the host veth name is %s", cfg.HostVETHName)
		hostVethLink, err := netlink.LinkByName(cfg.HostVETHName)
		if err != nil {
			log.Fatalf("error found link %s,%w", cfg.HostVETHName, err)
		}
		_, err = EnsureLinkUp(hostVethLink)
		if err != nil {
			log.Fatalf("error set link %s to up, %w", hostVethLink, err)
		}
		rate, err := Translate(cfg.Ingress)
		log.Printf("the rate translated is %d", rate)
		if err != nil {
			log.Fatalf("error get rate %s", err)
		}
		ReplaceTbf(hostVethLink, rate)
	}

	//set Egress

}

//EnsureLinkUp set link up, return changed and err
func EnsureLinkUp(link netlink.Link) (bool, error) {
	if link.Attrs().Flags&net.FlagUp != 0 {
		return false, nil
	}
	return true, LinkSetUp(link)
}

func LinkSetUp(link netlink.Link) error {
	errInfo := fmt.Sprintf("ip link set %s up", link.Attrs().Name)
	err := netlink.LinkSetUp(link)
	if err != nil {
		return fmt.Errorf("error %s,%w", errInfo, err)
	}
	return nil
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
