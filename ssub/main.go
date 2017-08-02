package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	proto "github.com/huin/mqtt"
	"github.com/ingemar0720/mqtt"
)

var host = flag.String("host", "localhost:8883", "hostname of broker")
var id = flag.String("id", "", "client id")
var user = flag.String("user", "", "username")
var pass = flag.String("pass", "", "password")
var dump = flag.Bool("dump", false, "dump messages?")

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: sub topic [topic topic...]")
		return
	}

	CA_Pool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile("./cert_pub.pem")
	if err != nil {
		log.Fatal("Could not load server certificate!")
	}
	CA_Pool.AppendCertsFromPEM(severCert)

	cert, err := tls.LoadX509KeyPair("./cert_sub.pem", "./key_sub.pem")
	if err != nil {
		log.Fatal("could not load client certificate!")
	}

	conf := &tls.Config{
		RootCAs:      CA_Pool,
		Certificates: []tls.Certificate{cert},
	}

	conf.BuildNameToCertificate()

	conn, err := tls.Dial("tcp", *host, conf)
	if err != nil {
		fmt.Fprint(os.Stderr, "dial: ", err)
		return
	}
	cc := mqtt.NewClientConn(conn)
	cc.Dump = *dump
	cc.ClientId = *id

	tq := make([]proto.TopicQos, flag.NArg())
	for i := 0; i < flag.NArg(); i++ {
		tq[i].Topic = flag.Arg(i)
		tq[i].Qos = proto.QosAtMostOnce
	}

	if err := cc.Connect(*user, *pass); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected with client id", cc.ClientId)
	cc.Subscribe(tq)

	for m := range cc.Incoming {
		fmt.Print(m.TopicName, "\t")
		m.Payload.WritePayload(os.Stdout)
		fmt.Println("\tr: ", m.Header.Retain)
	}
}
