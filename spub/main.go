package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	proto "github.com/huin/mqtt"
	"github.com/ingemar0720/mqtt"
	"github.com/ingemar0720/mqtt/lib"
)

var host = flag.String("host", "localhost:8883", "hostname of broker")
var user = flag.String("user", "", "username")
var pass = flag.String("pass", "", "password")
var dump = flag.Bool("dump", false, "dump messages?")
var retain = flag.Bool("retain", false, "retain message?")
var wait = flag.Bool("wait", true, "stay connected after publishing?")
var SBC_ID = "client0001"

var lastPublishedTime time.Time
var ch = make(chan int, 1)

func getCurrentTime() time.Time {
	lastPublishedTime = time.Now()
	return lastPublishedTime
}

func publish(cc *mqtt.ClientConn, topic string, payload []byte) {
	var tmpArray []byte
	tmpArray = append(tmpArray, payload...)
	cc.Publish(&proto.Publish{
		Header:    proto.Header{Retain: *retain},
		TopicName: topic,
		Payload:   proto.BytesPayload(tmpArray),
	})
}

func sendHeartbeat(cc *mqtt.ClientConn) {
	for {
		currTime := time.Now()
		if currTime.Sub(lastPublishedTime).Seconds() > 3.0 {
			var topic bytes.Buffer
			topic.WriteString(SBC_ID)
			topic.WriteString("/Heartbeat")
			heartbeatp := IoT.IoTPayload{DeviceID: SBC_ID, Timestamp: getCurrentTime()}
			b, err := heartbeatp.MarshalJSON()
			//b, err := json.Marshal(heartbeatp)
			if err == nil {
				<-ch
				publish(cc, topic.String(), b)
				ch <- 1
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func updateStatus(cc *mqtt.ClientConn) {
	for {
		var topic bytes.Buffer
		topic.WriteString(SBC_ID)
		topic.WriteString("/SolarStatus")

		solarp1 := IoT.IoTPayload{DeviceID: SBC_ID, Timestamp: getCurrentTime(), Irradiance: 99, Current: 99}
		b1, err := solarp1.MarshalJSON()
		if err == nil {
			<-ch
			publish(cc, topic.String(), b1)
			ch <- 1
		}
		time.Sleep(2 * time.Second)

		solarp2 := IoT.IoTPayload{DeviceID: SBC_ID, Timestamp: getCurrentTime(), Irradiance: 50, Current: 50}
		b2, err := solarp2.MarshalJSON()
		if err == nil {
			<-ch
			publish(cc, topic.String(), b2)
			ch <- 1
		}
		time.Sleep(6 * time.Second)
	}
}

func main() {
	flag.Parse()

	CA_Pool := x509.NewCertPool()
	severCert, err := ioutil.ReadFile("./cert_pub.pem")
	if err != nil {
		log.Fatal("Could not load server certificate!")
	}
	CA_Pool.AppendCertsFromPEM(severCert)

	cert, err := tls.LoadX509KeyPair("./cert_pub.pem", "./key_pub.pem")
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

	if err := cc.Connect(*user, *pass); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected with client id", cc.ClientId)
	ch <- 1
	//go sendHeartbeat(cc)
	go updateStatus(cc)

	if *wait {
		<-make(chan bool)
	}

	cc.Disconnect()
}
