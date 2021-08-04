package controller

import "github.com/containernetworking/plugins/pkg/ns"

type SetRuleConfig struct {
	Ingress      string
	Egress       string
	HostVethIndex int //set ingress rule on host veth
	ContVethIndex int //set egress rule on container veth
	HostNetwork  bool
	containerNetNs	ns.NetNS
}

//tc unit (byte)
const (
	TC_BPS = ""
	TC_KPS = "k"
	TC_MPS = "m"
	TC_GPS = "g"
	TC_TPS = "t"
)
