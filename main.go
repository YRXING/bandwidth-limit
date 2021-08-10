package main

import (
	"bandwidth-limit/controller"
	"flag"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"


)

func main() {
	var client *clientset.Clientset
	var err error
	var edge *bool
	var kubeconfig *string

	if home := homedir.HomeDir();home != ""{
		kubeconfig = flag.String("kubeconfig",filepath.Join(home,".kube","config"),"(optional absolute path to the kubeconfig file)")
	} else {
		kubeconfig = flag.String("kubeconfig","","absolute path to the kubeconfig file")
	}
	edge = flag.Bool("edge",false,"used in edge node or not, default is false.")

	flag.Parse()

	if *edge {
		client,err = clientset.NewForConfig(&rest.Config{
			Host: controller.Metaserver_Httpaddr,
		})
		if err != nil {
			panic(err)
		}
	}else{
		config,err := clientcmd.BuildConfigFromFlags("",*kubeconfig)
		if err != nil {
			panic(err)
		}
		client,err = clientset.NewForConfig(config)
		if err != nil {
			panic(err)
		}
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	sharedInformers := informers.NewSharedInformerFactory(client,0)
	podInformer := sharedInformers.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.AddPod,
		UpdateFunc: controller.UpdatePod,
		DeleteFunc: controller.DeletePod,
	})

	podInformer.Run(stopCh)
}
