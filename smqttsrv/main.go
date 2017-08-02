package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/ingemar0720/mqtt"
)

func readCert() []tls.Certificate {
	c, err := tls.LoadX509KeyPair("./cert_pub.pem", "./key_pub.pem")
	if err != nil {
		panic(err)
	}
	return []tls.Certificate{c}
}

func main() {

	certPubBytes, err := ioutil.ReadFile("./cert_pub.pem")
	if err != nil {
		log.Fatalln("Unable to read cert.pem", err)
	}

	certSubBytes, err := ioutil.ReadFile("./cert_sub.pem")
	if err != nil {
		log.Fatalln("Unable to read cert.pem", err)
	}

	clientCertPool := x509.NewCertPool()
	if ok := clientCertPool.AppendCertsFromPEM(certPubBytes); !ok {
		log.Fatalln("Unable to add publish certificate to certificate pool")
	}

	ok := clientCertPool.AppendCertsFromPEM(certSubBytes)
	if ok != true {
		log.Fatalln("Unable to add subscribe certificate to certificate pool")
	}

	cfg := &tls.Config{
		Certificates: readCert(),
		NextProtos:   []string{"mqtt"},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		ClientCAs: clientCertPool,
		// Force it server side
		PreferServerCipherSuites: true,
		// TLS 1.2 because we can
		MinVersion: tls.VersionTLS12,
	}
	l, err := tls.Listen("tcp", ":8883", cfg)
	if err != nil {
		log.Print("listen: ", err)
		return
	}
	svr := mqtt.NewServer(l)
	svr.Start()
	<-svr.Done
}
