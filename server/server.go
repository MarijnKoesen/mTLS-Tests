package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Hello from test server.\n"))
}

func handleError(err error) {
	if err != nil {
		log.Fatal("Fatal", err)
	}
}

func main() {
	serverCrt := os.Getenv("SERVER_CRT")
	if serverCrt == "" {
		serverCrt = "certs/server.crt"
	}

	serverKey := os.Getenv("SERVER_KEY")
	if serverKey == "" {
		serverKey = "certs/server.key"
	}

	caCrt := os.Getenv("CA_CRT")
	if caCrt == "" {
		caCrt = "certs/server.crt"
	}

	absPathServerCrt, err := filepath.Abs(serverCrt)
	handleError(err)
	absPathServerKey, err := filepath.Abs(serverKey)
	handleError(err)

	clientCACert, err := ioutil.ReadFile(caCrt)
	handleError(err)

	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientCACert)

	tlsConfig := &tls.Config{
		ClientAuth:               tls.RequireAndVerifyClientCert,
		ClientCAs:                clientCertPool,
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
	}

	http.HandleFunc("/", HelloServer)
	httpServer := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
	}

	fmt.Println("Running Server...")
	err = httpServer.ListenAndServeTLS(absPathServerCrt, absPathServerKey)
	handleError(err)
}
