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

func NewMqttClient(mqttIp string, mqttPort string, mqttClientId string) *MqttClient {
	return &MqttClient{
		client: connect(mqttIp, mqttPort, mqttClientId),
	}
}

func connect(mqttIp string, mqttPort string, clientId string) mqtt.Client {
	opts := createClientOptions(clientId, fmt.Sprintf("tcp://%s:%s", mqttIp, mqttPort))
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Println(err)
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

	token := c.client.Publish(topic, 0, false, body)
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Println(err)
	}
}

func (c *MqttClient) IsConnected() bool {
	return c.client.IsConnectionOpen()
}
