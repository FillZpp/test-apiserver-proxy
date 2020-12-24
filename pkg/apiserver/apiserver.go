package apiserver

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"k8s.io/klog"

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

func StartHTTPS(stopCh <-chan struct{}) error {
	sv := &http.Server{
		Addr: ":6443",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				klog.Info("Handle tunneling")
				handleTunneling(w, r)
			} else {
				klog.Info("Handle HTTP")
				handleHTTP(w, r)
			}
		}),
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	go func() {
		if err := sv.ListenAndServeTLS("./cert/cert.pem", "./cert/key.pem"); err != nil {
			klog.Fatalf("Failed to start HTTPS: %v", err)
		}
	}()
	return nil
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	dest_conn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)
}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}
