package IoT

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huin/mqtt"
)

var m map[string]time.Time = make(map[string]time.Time)
var mutex = &sync.Mutex{}

type IoTPayload struct {
	Timestamp  time.Time `json:"timestamp"`
	DeviceID   string    `json:"Client"`
	Irradiance int       `json:"irradiance"`
	Current    int       `json:"current"`
}

func (p *IoTPayload) UnmarshalJSON(j []byte) error {
	var rawStrings map[string]string

	err := json.Unmarshal(j, &rawStrings)
	if err != nil {
		fmt.Println("fail unmarshal to rawsstrings")
		return err
	}

	for k, v := range rawStrings {
		if strings.ToLower(k) == "deviceid" {
			p.DeviceID = v
		} else if strings.ToLower(k) == "timestamp" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				fmt.Println("fail unmarshall timestamp")
				return err
			}
			p.Timestamp = t
		} else if strings.ToLower(k) == "irradiance" {
			p.Irradiance, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		} else if strings.ToLower(k) == "current" {
			p.Current, err = strconv.Atoi(v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *IoTPayload) MarshalJSON() ([]byte, error) {
	newPayload := struct {
		DeviceID   string `json:"deviceid"`
		Timestamp  string `json:"timestamp"`
		Irradiance string `json:"irradiance"`
		Current    string `json:"current"`
	}{
		DeviceID:   p.DeviceID,
		Timestamp:  p.Timestamp.Format(time.RFC3339),
		Irradiance: strconv.Itoa(p.Irradiance),
		Current:    strconv.Itoa(p.Current),
	}
	return json.Marshal(newPayload)
}

func getCurrentTime() time.Time {
	currTime := time.Now()
	return currTime
}

func CheckAlive() {
	for {
		count := 0
		newTime := getCurrentTime()
		mutex.Lock()
		for k, v := range m {
			duration := newTime.Sub(v).Seconds()
			if duration > 4.0 {
				fmt.Printf("Device %s is disconnected\n", k)
			}
			count++
		}
		mutex.Unlock()
		time.Sleep(time.Second * 1)
	}
}

func updateTimeMap(message IoTPayload) {
	_, ok := m[message.DeviceID]
	if ok != true {
		fmt.Printf("%s, %s\n", "register new device", message.DeviceID)
	} else {
		fmt.Println("update device ", message.DeviceID, " timestamp ", message.Timestamp)
	}
	m[message.DeviceID] = message.Timestamp
}

func RegisterSBC(topic string, payload mqtt.Payload) {
	var message IoTPayload

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, payload)

	err := message.UnmarshalJSON(buf.Bytes())
	if err != nil {
		fmt.Printf("%s %s\n", "Unmarshal JSON fail", err)
		return
	}
	mutex.Lock()
	updateTimeMap(message)
	mutex.Unlock()
}
