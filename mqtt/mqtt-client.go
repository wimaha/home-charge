package mqtt

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient struct {
	client mqtt.Client
}

func NewMqttClient(mqttIp string, mqttPort string) *MqttClient {
	return &MqttClient{
		client: connect(mqttIp, mqttPort),
	}
}

func connect(mqttIp string, mqttPort string) mqtt.Client {
	opts := createClientOptions("homeCharge", fmt.Sprintf("tcp://%s:%s", mqttIp, mqttPort))
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, url string) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(url)
	opts.SetClientID(clientId)
	return opts
}

func (c *MqttClient) Publish(topic string, body string) {
	if !c.client.IsConnected() {
		c.client.Connect()
	}

	c.client.Publish(topic, 0, false, body)
}
