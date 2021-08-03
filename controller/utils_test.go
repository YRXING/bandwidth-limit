package controller

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
	"testing"
)

func TestSetTcRule(t *testing.T) {
	var ingress, hostVethName string
	fmt.Println("print ingress:")
	fmt.Scanf("%s", &ingress)
	fmt.Println("print dev name:")
	fmt.Scanf("%s", &hostVethName)
	log.Printf("you want to limit %s to dev %s", ingress, hostVethName)
	cfg := &SetRuleConfig{
		Ingress:      ingress,
		HostVETHName: hostVethName,
		HostNetwork:  false,
	}
	SetTcRule(cfg)
	link, _ := netlink.LinkByName(hostVethName)
	fmt.Println(netlink.QdiscList(link))
}
