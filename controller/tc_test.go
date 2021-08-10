package controller

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"testing"
)

func TestSetTcRule(t *testing.T) {
	var ingress string
	var hostVethIndex int
	fmt.Println("print ingress:")
	fmt.Scanf("%s", &ingress)
	fmt.Println("print dev index:")
	fmt.Scanf("%d", &hostVethIndex)
	log.Printf("you want to limit %s to dev %d", ingress, hostVethIndex)
	cfg := &SetRuleConfig{
		Ingress:      ingress,
		HostVethIndex: hostVethIndex,
		HostNetwork:  false,
	}
	SetTcRule(cfg)
	link, _ := netlink.LinkByIndex(hostVethIndex)
	fmt.Println(netlink.QdiscList(link))
}
