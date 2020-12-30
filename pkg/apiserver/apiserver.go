package apiserver

import (
	"net/http"
	"time"

	"github.com/jiuzhu.wsy/test-apiserver-proxy/pkg/handler"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	cfg       *rest.Config
	transport http.RoundTripper
)

func init() {
	var err error
	cfg = config.GetConfigOrDie()
	if transport, err = rest.TransportFor(cfg); err != nil {
		panic(err)
	}
}

func Start(stopCh <-chan struct{}) error {
	opts := options.NewSecureServingOptions()
	opts.ServerCert.CertKey.CertFile = "./cert/cert.pem"
	opts.ServerCert.CertKey.KeyFile = "./cert/key.pem"
	opts.BindPort = 6443
	var apiserver *server.SecureServingInfo
	if err := opts.ApplyTo(&apiserver); err != nil {
		return err
	}
	_, err := apiserver.Serve(handler.NewResourceHandler(cfg, transport), time.Minute, stopCh)
	if err != nil {
		return err
	}
	return nil
}
