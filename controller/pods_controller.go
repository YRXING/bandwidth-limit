package controller

import (

	v1 "k8s.io/api/core/v1"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"os"
	"time"
)

type PodController struct {
	client        clientset.Interface
	eventRecorder record.EventRecorder
	podLister     corelisters.PodLister
	podsSynced    cache.InformerSynced
	cfg           *SetRuleConfig
}

func NewPodController(podInformer coreinformers.PodInformer, client clientset.Interface) *PodController {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{
		Interface: client.CoreV1().Events(""),
	})

	pc := &PodController{
		client:        client,
		eventRecorder: eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "pods-controller"}),
	}
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{

	})
	pc.podLister = podInformer.Lister()
	pc.podsSynced = podInformer.Informer().HasSynced

	return pc
}

// Run will not return until stopCh is closed. workers determines how many
// pods will be handled in parallel.
func (pc *PodController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	klog.Infof("Starting pods controller")
	defer klog.Infof("Shutting down pods controller")

	if !cache.WaitForNamedCacheSync("pods", stopCh, pc.podsSynced) {
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(pc.worker, time.Second, stopCh)
	}

	<-stopCh

}

func (pc *PodController) worker() {

}

func AddPod(obj interface{}) {
	pod := obj.(*v1.Pod)
	hostName , err := os.Hostname()
	if err != nil {
		klog.Info(err)
	}
	if hostName != pod.Spec.NodeName {
		return
	}

	cfg := &SetRuleConfig{
		Ingress:     pod.Annotations["k8s.harmonycloud.com/ingress-bandwidth"],
		Egress:      pod.Annotations["k8s.harmonycloud.com/egress-bandwidth"],
		HostNetwork: pod.Spec.HostNetwork,
		HostVethIndex: -1,
		ContVethIndex: -1,
	}

	if cfg.Ingress != "" || cfg.Egress != "" {
		if pod.Status.ContainerStatuses == nil {
			klog.Infof("Add event: container creating...")
		}else if len(pod.Status.ContainerStatuses[0].ContainerID) == 0 {
			klog.Infof("Add evnet: container creating...")
		}else{
			klog.Infof("Add event: container exists ")
			containerId := pod.Status.ContainerStatuses[0].ContainerID[9:]
			klog.Infof("pod's contaier id is: %+v",containerId)
			containerPid := GetContainerPid(containerId)
			if containerPid == "0" {
				klog.Errorf("container is not running")
				return
			}
			klog.Infof("pod's container pid is: %s",containerPid)
			//ExposeNetNs(containerPid)
			GetVethInfo(containerPid,cfg)
			klog.Infof("Tc set up rules: %+v",cfg)
			if cfg.containerNetNs != nil {
				klog.Info("start setting tc rules....")
				SetTcRule(cfg)
			}else {
				klog.Info("missing necessary information, tc rules set failed!")
			}
		}
	}

}

// Scale product update event even if you scale pod to 0
// When pod on this node leaves, it product update event first, and delete event in the end.
func UpdatePod(oldObj, newObj interface{}) {
	pod := newObj.(*v1.Pod)
	hostName , err := os.Hostname()
	if err != nil {
		klog.Info(err)
	}
	if hostName != pod.Spec.NodeName {
		return
	}
	cfg := &SetRuleConfig{
		Ingress:     pod.Annotations["k8s.harmonycloud.com/ingress-bandwidth"],
		Egress:      pod.Annotations["k8s.harmonycloud.com/egress-bandwidth"],
		HostNetwork: pod.Spec.HostNetwork,
		HostVethIndex: -1,
		ContVethIndex: -1,
	}

	if cfg.Ingress != "" || cfg.Egress != "" {
		if pod.Status.ContainerStatuses == nil {
			klog.Infof("Update event: container creating...")
		}else if len(pod.Status.ContainerStatuses[0].ContainerID) == 0 {
			klog.Infof("Update evnet: container creating...")
		}else{
			klog.Infof("Update event: container exists ")
			containerId := pod.Status.ContainerStatuses[0].ContainerID[9:]
			klog.Infof("pod's contaier id is: %+v",containerId)
			containerPid := GetContainerPid(containerId)
			if containerPid == "0" {
				klog.Errorf("container is not running")
				return
			}
			klog.Infof("pod's container pid is: %s",containerPid)
			//ExposeNetNs(containerPid)
			GetVethInfo(containerPid,cfg)
			klog.Infof("Tc set up rules: %+v",cfg)
			if cfg.containerNetNs != nil {
				klog.Info("start setting tc rules....")
				SetTcRule(cfg)
			}else {
				klog.Info("missing necessary information, tc rules set failed!")
			}
		}
	}

}

func DeletePod(obj interface{}) {
	klog.Info("Delete event:...")
	//TODO: tear down
	//pod := obj.(*v1.Pod)
	//hostName , err := os.Hostname()
	//if err != nil {
	//	klog.Info(err)
	//}
	//if hostName != pod.Spec.NodeName {
	//	return
	//}
	//
	//cfg := &SetRuleConfig{
	//	Ingress:     pod.Annotations["ingress-bandwidth"],
	//	Egress:      pod.Annotations["egress-bandwidth"],
	//	HostNetwork: pod.Spec.HostNetwork,
	//	HostVethIndex: -1,
	//	ContVethIndex: -1,
	//}
	//if cfg.Ingress != "" || cfg.Egress != "" {
	//	containerId := pod.Status.ContainerStatuses[0].ContainerID[9:]
	//	klog.Info(containerId)
	//	containerPid := GetContainerPid(containerId)
	//	ExposeNetNs(containerPid)
	//	GetVethInfo(containerPid,cfg)
	//	klog.Info(cfg)
	//	SetTcRule(cfg)
	//}

}
