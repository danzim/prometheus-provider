package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danzim/prometheus-provider/pkg/handler"
	"github.com/danzim/prometheus-provider/pkg/utils"
	"k8s.io/klog/v2"
)

var (
	certDir      string
	clientCAFile string
	portHTTP     int
	portHTTPS    int
	useHTTPS     bool
)

const (
	timeout    = 3 * time.Second
	apiVersion = "externaldata.gatekeeper.sh/v1alpha1"
)

/* func init() {
	klog.InitFlags(nil)
	flag.BoolVar(&useHTTPS, "https", false, "Start the server with HTTPS")
	flag.StringVar(&certDir, "cert-dir", "", "path to directory containing TLS certificates")
	flag.StringVar(&clientCAFile, "client-ca-file", "", "path to client CA certificate")
	flag.IntVar(&portHTTP, "defaultPortHTTP", defaultPortHTTP, "Port for the server to listen on")
	flag.IntVar(&portHTTPS, "defaultPortHTTPS", defaultPortHTTPS, "Port for the server to listen on")
	flag.Parse()
} */

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", processTimeout(handler.Handler, timeout))

	if utils.AppConfig.Server.UseHTTPS {
		klog.InfoS("starting server with HTTPS...")
		server := &http.Server{
			Addr:              fmt.Sprintf(":%d", utils.AppConfig.Server.Port),
			Handler:           mux,
			ReadHeaderTimeout: time.Duration(5) * time.Second,
		}

		configTLS := &tls.Config{
			MinVersion: tls.VersionTLS13,
		}

		if utils.AppConfig.Server.ClientCAFile != "" {
			klog.InfoS("loading CA certificate", "clientCAFile", utils.AppConfig.Server.ClientCAFile)
			caCert, err := os.ReadFile(utils.AppConfig.Server.ClientCAFile)
			if err != nil {
				klog.ErrorS(err, "unable to load CA certificate", "clientCAFile", utils.AppConfig.Server.ClientCAFile)
				os.Exit(1)
			}

			clientCAs := x509.NewCertPool()
			clientCAs.AppendCertsFromPEM(caCert)

			configTLS.ClientCAs = clientCAs
			configTLS.ClientAuth = tls.RequireAndVerifyClientCert
			server.TLSConfig = configTLS

			if utils.AppConfig.Server.KeyFile != "" || utils.AppConfig.Server.CertFile != "" {

				klog.InfoS("starting external data provider server", "port", utils.AppConfig.Server.Port, "certFile", utils.AppConfig.Server.CertFile, "keyFile", utils.AppConfig.Server.KeyFile)
				if err := server.ListenAndServeTLS(utils.AppConfig.Server.CertFile, utils.AppConfig.Server.KeyFile); err != nil {
					klog.ErrorS(err, "unable to start external data provider server")
					os.Exit(1)
				}
			} else {
				klog.Error("TLS certificates are not provided, the server will not be started")
				os.Exit(1)
			}
		}
	} else {
		server := &http.Server{
			Addr:              fmt.Sprintf(":%d", utils.AppConfig.Server.Port),
			Handler:           mux,
			ReadHeaderTimeout: time.Duration(5) * time.Second,
		}
		klog.InfoS("starting external data provider server", "port", utils.AppConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil {
			klog.ErrorS(err, "unable to start external data provider server")
			os.Exit(1)
		}

	}

}

func processTimeout(h http.HandlerFunc, duration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), duration)
		defer cancel()

		r = r.WithContext(ctx)

		processDone := make(chan bool)
		go func() {
			h(w, r)
			processDone <- true
		}()

		select {
		case <-ctx.Done():
			utils.SendResponse(nil, "operation timed out", w)
		case <-processDone:
		}
	}
}
