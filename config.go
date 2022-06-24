package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//Config global
var Config = loadConfig()

//ServerST struct
type ConfigST struct {
	ImgFileName  string `json:"imgFileName"`
	MQTTserver   string `json:"mqttServer"`
	MQTTuserName string `json:"mqttUserName"`
	MQTTpassword string `json:"mqttPassword"`
}

//读取配置文件并生成附属结构
func loadConfig() *ConfigST {
	var tmp ConfigST
	data, err := ioutil.ReadFile("config.json")
	if err == nil {
		err = json.Unmarshal(data, &tmp)
		if err != nil {
			log.Fatalln(err)
		}

	} else {
		log.Println("read config files wrong:", err)
	}
	return &tmp
}
