package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
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
	port         int
)

const (
	timeout    = 3 * time.Second
	apiVersion = "externaldata.gatekeeper.sh/v1alpha1"

	defaultPort = 8080

	certName = "tls.crt"
	keyName  = "tls.key"
)

func init() {
	klog.InitFlags(nil)
	flag.StringVar(&certDir, "cert-dir", "", "path to directory containing TLS certificates")
	flag.StringVar(&clientCAFile, "client-ca-file", "", "path to client CA certificate")
	flag.IntVar(&port, "defaultPort", defaultPort, "Port for the server to listen on")
	flag.Parse()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", processTimeout(handler.Handler, timeout))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: time.Duration(5) * time.Second,
	}

	config := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	if clientCAFile != "" {
		klog.InfoS("loading CA certificate", "clientCAFile", clientCAFile)
		caCert, err := os.ReadFile(clientCAFile)
		if err != nil {
			klog.ErrorS(err, "unable to load CA certificate", "clientCAFile", clientCAFile)
			os.Exit(1)
		}

		clientCAs := x509.NewCertPool()
		clientCAs.AppendCertsFromPEM(caCert)

		config.ClientCAs = clientCAs
		config.ClientAuth = tls.RequireAndVerifyClientCert
		server.TLSConfig = config
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
