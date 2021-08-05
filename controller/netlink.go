package controller

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net"
	"os"
	"path/filepath"
	"strings"
	"github.com/containernetworking/plugins/pkg/ns"
)


func LimitPort(dev netlink.Link) {

}

func LimitIP() {

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


// Read "/var/lib/docker/containers/container_id/config.json" to get container's information.
func GetContainerPid(containerId string) string{
	path := filepath.Join("/var/lib/docker/containers/",containerId,"config.v2.json")

	file,err := os.Open(path)
	defer file.Close()
	if err != nil {
		klog.Error(err)
	}

	//read all content
	data,err := ioutil.ReadAll(file)
	if err != nil {
		klog.Error(err)
	}

	pid := GetFiledFromJason(strings.Split("State/Pid","/"),data)

	return pid
}

// Because docker hides the created network namespace link file to "/var/run/docker/netns".
// And ip netns command only find the net namespace in "/var/run/netns", so we should restore association.
// But if we only want to enter the container's network namespace, just open "/proc/pid/ns/net" file.
// Equivalent to: `ln -sf "/proc/pid/ns/net" "/var/run/netns/container_id"`
func ExposeNetNs(pid string){
	//create symbolic link to container's process net namespace
	netFile := filepath.Join("/proc",pid,"ns/net")
	symlinkNetFile := filepath.Join("/var/run/netns/ns-"+pid)

	//check if netns directory exists
	//if not, create it.
	_,err := os.Stat("/var/run/netns")
	if err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("/var/run/netns", 0777)
		}
	}

	err = os.Symlink(netFile,symlinkNetFile)
	if err != nil {
		klog.Error(err)
		if os.IsExist(err) {
			os.Remove(symlinkNetFile)
			os.Symlink(netFile,symlinkNetFile)
		}
	}

}

// Enter container's net namespace to get containeVeth and hostVeth's index
func GetVethInfo(containerPid string,cfg *SetRuleConfig){

	netFile := filepath.Join("/proc",containerPid,"ns/net")
	contaienrNetNs , err := ns.GetNS(netFile)
	if err != nil {
		klog.Error(err)
	}

	cfg.containerNetNs = contaienrNetNs

	err = contaienrNetNs.Do(func(_ ns.NetNS) error {
		links,err:= netlink.LinkList()
		if err != nil {
			klog.Error(err)
			return err
		}

		for _,l := range links{
			if l.Type()=="veth"{
				veth := l.(*netlink.Veth)
				cfg.ContVethIndex = veth.Index
				cfg.HostVethIndex, err = netlink.VethPeerIndex(veth)
				if err != nil {
					klog.Info(err)
					return err
				}
				break
			}
		}
		return nil
	})

}

func GetHostLink() netlink.Link {
	return nil
}
