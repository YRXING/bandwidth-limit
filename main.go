package main

import (
	"bandwidth-limit/controller"
	"fmt"
	"github.com/vishvananda/netlink"
	"log"
)

func Pow(data,n int) uint64 {
	for i:=0;i<n;i++ {
		data*=10
	}
	return uint64(data)
}
func main() {
	var ingress,hostVethName string
	fmt.Println("print ingress:")
	fmt.Scanf("%s",&ingress)
	fmt.Println("print dev name:")
	fmt.Scanf("%s",&hostVethName)
	log.Printf("you want to limit %s to dev %s",ingress,hostVethName)
	cfg := &controller.SetRuleConfig{
		Ingress: ingress,
		HostVETHName: hostVethName,
		HostNetwork: false,
	}
	controller.SetTcRule(cfg)

	link, _ := netlink.LinkByName(hostVethName)
	fmt.Println(netlink.QdiscList(link))
}