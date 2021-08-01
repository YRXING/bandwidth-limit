package controller

type SetRuleConfig struct {
	Ingress string
	Egress string
	HostVETHName string //set ingress rule on host veth
	ContVethName string //set egress rule on container veth
	HostNetwork bool
}

