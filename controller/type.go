package controller

type SetRuleConfig struct {
	Ingress string
	Egress string
	HostVETHName string //set ingress rule on host veth
	ContVethName string //set egress rule on container veth
	HostNetwork bool
}

//tc unit (byte)
const (
	TC_BPS	=	""
	TC_KPS	=	"k"
	TC_MPS	=	"m"
	TC_GPS	=	"g"
	TC_TPS	=	"t"
)
