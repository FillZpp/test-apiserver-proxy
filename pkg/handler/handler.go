package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/emicklei/go-restful"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

type resourceHandler struct {
	cfg       *rest.Config
	transport http.RoundTripper
	container *restful.Container
}

func NewResourceHandler(cfg *rest.Config, transport http.RoundTripper) *resourceHandler {
	h := &resourceHandler{container: restful.NewContainer(), cfg: cfg, transport: transport}
	//h.Install()
	return h
}

func (h *resourceHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	klog.Infof("%s %s %s", r.Method, r.Header.Get("Content-Type"), r.URL)
	//h.container.Dispatch(rw, r)
	h.proxyHandler(rw, r)
}

func (h *resourceHandler) Install() {
	h.container.ServiceErrorHandler(func(err restful.ServiceError, r *restful.Request, rw *restful.Response) {
		if err.Code != 404 {
			http.Error(rw.ResponseWriter, err.Message, err.Code)
			return
		}
		h.proxyHandler(rw.ResponseWriter, r.Request)
	})
	//handler.Install(h.container)
}

func (h *resourceHandler) proxyHandler(rw http.ResponseWriter, r *http.Request) {
	url, _ := url.Parse(h.cfg.Host)
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = h.transport
	// TODO: set labels
	proxy.ServeHTTP(rw, r)
}
