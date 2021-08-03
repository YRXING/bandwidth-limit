package controller

import "github.com/vishvananda/netlink"

func LimitVeth(veth netlink.Veth) {

}
func LimitPort(dev netlink.Link) {}

func LimitIP() {
}

func GetHostLink() netlink.Link {
	return nil
}
