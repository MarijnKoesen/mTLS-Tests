package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func handleError(err error) {
	if err != nil {
		log.Fatal("Fatal", err)
	}
}

func main() {
	clientCrt := os.Getenv("CLIENT_CRT")
	if clientCrt == "" {
		clientCrt = "certs/client.crt"
	}

	clientKey := os.Getenv("CLIENT_KEY")
	if clientKey == "" {
		clientKey = "certs/client.key"
	}

	caPath := os.Getenv("CA_CRT")
	if caPath == "" {
		caPath = "certs/server.crt"
	}

	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = "https://localhost"
	}

	absPathClientCrt, err := filepath.Abs(clientCrt)
	handleError(err)
	absPathClientKey, err := filepath.Abs(clientKey)
	handleError(err)
	caCrt, err := filepath.Abs(caPath)
	handleError(err)

	cert, err := tls.LoadX509KeyPair(absPathClientCrt, absPathClientKey)
	if err != nil {
		log.Fatalln("Unable to load cert", err)
	}

	fmt.Printf("Loaded certificate: %s\n", base64.StdEncoding.EncodeToString(cert.Certificate[0]))

	roots := x509.NewCertPool()

	// We're going to load the server cert and add all the intermediates and CA from that.
	// Alternatively if we have the CA directly we could call AppendCertificate method
	fakeCA, err := ioutil.ReadFile(caCrt)
	if err != nil {
		log.Println(err)
		return
	}

	ok := roots.AppendCertsFromPEM([]byte(fakeCA))
	if !ok {
		panic("failed to parse root certificate")
	}

	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            roots,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}
	tr := &http.Transport{TLSClientConfig: tlsConf}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(resp.Status)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(body))
}
