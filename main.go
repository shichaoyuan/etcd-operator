package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	clientset "etcd-operator/pkg/generated/clientset/versioned"
	informers "etcd-operator/pkg/generated/informers/externalversions"
	"etcd-operator/pkg/controller"
	"etcd-operator/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(client, time.Second*30)

	ctl := controller.NewController(kubeClient, client, informerFactory.Samplecrd().V1().EtcdClusters())

	go informerFactory.Start(stopCh)

	if err = ctl.Run(stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}

}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
