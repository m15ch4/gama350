package mqtt

import (
	"encoding/json"

	"micze.io/gama350/config"
	"micze.io/gama350/influx"
	"micze.io/gama350/logger"
	"micze.io/gama350/model"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client mqtt.Client
}

func NewClient(influxWriter influx.Writer) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(config.Get("MQTT_BROKER"))
	opts.SetClientID("mqtt-influx-writer")
	opts.SetUsername(config.Get("MQTT_USERNAME"))
	opts.SetPassword(config.Get("MQTT_PASSWORD"))
	opts.SetAutoReconnect(true)
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		logger.Logger.Printf("MQTT connection lost: %v", err)
	}
	opts.OnConnect = func(c mqtt.Client) {
		logger.Logger.Println("MQTT connected")
		c.Subscribe(config.Get("MQTT_TOPIC"), 0, func(client mqtt.Client, msg mqtt.Message) {
			var data model.MeterData
			if err := json.Unmarshal(msg.Payload(), &data); err != nil {
				logger.Logger.Printf("JSON parse error: %v", err)
				return
			}
			influxWriter.Write(data)
		})
	}

	client := mqtt.NewClient(opts)
	return &MQTTClient{client}
}

func (m *MQTTClient) Start() {
	token := m.client.Connect()
	token.Wait()
	if token.Error() != nil {
		logger.Logger.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}
}

func (m *MQTTClient) Stop() {
	m.client.Disconnect(250)
}
