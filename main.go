package main

import (
	"bandwidth-limit/controller"
	"flag"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir();home != ""{
		kubeconfig = flag.String("kubeconfig",filepath.Join(home,".kube","config"),"(optional absolute path to the kubeconfig file)")
	} else {
		kubeconfig = flag.String("kubeconfig","","absolute path to the kubeconfig file")
	}
	flag.Parse()
	config,err := clientcmd.BuildConfigFromFlags("",*kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset,err := clientset.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformers := informers.NewSharedInformerFactory(clientset,time.Minute)
	podInformer := sharedInformers.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.AddPod,
		UpdateFunc: controller.UpdatePod,
		DeleteFunc: controller.DeletePod,
	})

	podInformer.Run(stopCh)
}
