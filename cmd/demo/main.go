package main

import (
	"flag"
	"fmt"
	"k8s-code-gen-demo/generated/clientset/versioned"
	"k8s-code-gen-demo/generated/clientset/versioned/typed/democontroller/v1alpha1"
	"k8s-code-gen-demo/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

var log = logf.Log.WithName("cmd")

func main() {
	flag.Parse()
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	client, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	demoList, err := client.Demos("test").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	log.Info(fmt.Sprintf("demoList: [%s]", demoList))
	clientset, err := versioned.NewForConfig(cfg)
	if err != nil {
		panic(err.Error())
	}
	factory := externalversions.NewSharedInformerFactory(clientset, 30*time.Second)
	demo, err := factory.Democontroller().V1alpha1().Demos().Lister().Demos("test").Get("test")
	if err != nil {
		panic(err.Error())
	}
	log.Info(fmt.Sprintf("demo: [%s]", demo))
}
