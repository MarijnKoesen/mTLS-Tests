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
	"strings"
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

	// When ALLOWED_CLIENTS is specified only clients in this list will be allowed
	// Format is a list of emails split by comma, e.g. 'client1@example.com,client2@example.com'
	allowedClients := make([]string, 0)
	allowedClientsString := os.Getenv("ALLOWED_CLIENTS")
	if allowedClientsString != "" {
		allowedClients = strings.Split(allowedClientsString, ",")
		for i, _ := range allowedClients {
			allowedClients[i] = strings.TrimSpace(allowedClients[i])
		}
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
		VerifyConnection: func(state tls.ConnectionState) error {
			if len(allowedClients) == 0 {
				// No authentication
				return nil
			}

			clientEmails := make([]string, 0)
			for _, cert := range state.PeerCertificates {
				for _, clientEmail := range cert.EmailAddresses {
					clientEmails = append(clientEmails, clientEmail)
					if indexOf(clientEmail, allowedClients) > -1 {
						return nil
					}
				}
			}

			return fmt.Errorf("Client '" + strings.Join(clientEmails, ", ") + "' is not in allow list: " + strings.Join(allowedClients, ", "))
		},
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

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
