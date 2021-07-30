package controller

import (
"fmt"
"github.com/pkg/errors"
"github.com/vishvananda/netlink"
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

//tc qdisc replace dev $dev root tbf rate $rate latency 50ms burst 20k
func ReplaceTbf(dev netlink.Link, rate uint64) error {
	if rate <= 0 {
		return fmt.Errorf("invalid rate #{rate}")
	}
	burst := burst(rate, dev.Attrs().MTU+hardwareHeaderLen)
	buffer := buffer(rate,burst)
	latency := latencyInUsec(latencyInMillis)
	limit := limit(rate,latency,buffer)


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
func DeletRule(dev netlink.Link) error {
	//TODO
	return nil
}
