package main

import (
	"fmt"
	"log"

	"github.com/yosssi/gmq/mqtt/client"
)

// 建立到MQTT代理服务器的连接
func createMQTTClient(brokerAddr string, clientID string, userName string, password string) (cli *client.Client) {

	// Create an MQTT Client.
	cli = client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			log.Println("mqtt 生成错误 :", err)
		},
	})

	// Connect to the MQTT Server.
	err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  brokerAddr,
		UserName: []byte(userName),
		Password: []byte(password),
		ClientID: []byte(clientID),
	})

	if err != nil {
		fmt.Println("mqtt 连接错误 :", "服务器地址：", brokerAddr, err)
		panic(err)
	}
	return
}

// 在一个连接中发布主题消息
func publishMQTTtopic(mqttClietn *client.Client, topicName string, message string, qos byte) {
	err := mqttClietn.Publish(&client.PublishOptions{
		QoS:       qos,
		TopicName: []byte(topicName),
		Message:   []byte(message),
	})
	if err != nil {
		log.Println(err)
	}
}

// 订阅一个主题, 不要调用这个函数, 这里主要用来展现写法, 调用时自己编写
func subscribeMQTTtopic(mqttClient *client.Client, topicName string, qos byte) {
	// 订阅主题
	err := mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte(topicName),
				QoS:         qos,
				// Define the processing of the message handler.
				Handler: func(topicName11, message []byte) {
					fmt.Println(string(topicName11), string(message))
				},
			},
			&client.SubReq{
				TopicFilter: []byte("bar/#"),
				QoS:         qos,
				Handler: func(topicName22, message []byte) {
					fmt.Println(string(topicName22), string(message))
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

// 取消订阅一个主题
func unSubscribeMQTTtopic(mqttClietn *client.Client, topicName string) {
	// 订阅主题

	// 取消订阅.
	err := mqttClietn.Unsubscribe(&client.UnsubscribeOptions{
		TopicFilters: [][]byte{
			[]byte(topicName),
		},
	})
	if err != nil {
		panic(err)
	}
}
