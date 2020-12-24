package main

import (
	"github.com/jiuzhu.wsy/test-apiserver-proxy/pkg/apiserver"
	"k8s.io/klog"
)

func main() {
	stopCh := make(chan struct{})
	if err := apiserver.Start(stopCh); err != nil {
		klog.Errorf("Failed to start: %v", err)
		return
	}
	klog.Infof("apiserver-proxy started")
	<-stopCh
}
