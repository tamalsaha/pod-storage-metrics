package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func NewClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	ctrl.SetLogger(klogr.New())
	cfg := ctrl.GetConfigOrDie()
	cfg.QPS = 100
	cfg.Burst = 100

	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	return client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		//Opts: client.WarningHandlerOptions{
		//	SuppressWarnings:   false,
		//	AllowDuplicateLogs: false,
		//},
	})
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	kc, err := NewClient()
	if err != nil {
		return err
	}

	var pods core.PodList
	err = kc.List(context.TODO(), &pods, client.InNamespace("default"))
	if err != nil {
		panic(err)
	}
	for _, pod := range pods.Items {
		req := resource.Quantity{Format: resource.BinarySI}
		limit := resource.Quantity{Format: resource.BinarySI}
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil {
				var pvc core.PersistentVolumeClaim
				if err := kc.Get(context.TODO(), client.ObjectKey{Namespace: pod.Namespace, Name: vol.PersistentVolumeClaim.ClaimName}, &pvc); err == nil {
					req.Add(*pvc.Spec.Resources.Requests.Storage())
					limit.Add(*pvc.Status.Capacity.Storage())
				}
			}
		}

		fmt.Printf("--------------%s---------------\n", pod.Name)
		fmt.Println("req: ", req.String())
		fmt.Println("cap: ", limit.String())
	}

	return nil
}
