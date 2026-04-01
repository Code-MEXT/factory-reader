package reader

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTReader struct {
	Host   string
	Port   int
	Topic  string
	client mqtt.Client
	mu     sync.Mutex
	last   any
}

func NewMQTTReader(host string, port int, topic string) *MQTTReader {
	return &MQTTReader{Host: host, Port: port, Topic: topic}
}

func (r *MQTTReader) Connect() error {
	broker := fmt.Sprintf("tcp://%s:%d", r.Host, r.Port)
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("factory-reader-%d", time.Now().UnixNano())).
		SetConnectTimeout(5 * time.Second)

	r.client = mqtt.NewClient(opts)
	token := r.client.Connect()
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("mqtt connect timeout to %s", broker)
	}
	if token.Error() != nil {
		return fmt.Errorf("mqtt connect error: %w", token.Error())
	}

	r.client.Subscribe(r.Topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.last = string(msg.Payload())
	})

	return nil
}

func (r *MQTTReader) Read() (*Result, error) {
	res := &Result{
		Protocol:  "mqtt",
		Host:      r.Host,
		Port:      r.Port,
		Connected: r.client != nil && r.client.IsConnected(),
		Timestamp: time.Now(),
	}
	if !res.Connected {
		res.Error = "not connected"
		return res, fmt.Errorf("mqtt not connected")
	}
	r.mu.Lock()
	res.Data = r.last
	r.mu.Unlock()
	return res, nil
}

func (r *MQTTReader) Close() error {
	if r.client != nil && r.client.IsConnected() {
		r.client.Disconnect(250)
	}
	return nil
}
