package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// 生成配置文件结构
// global config
var config = loadConfig()

type configST struct {
	Streams map[string]StreamST `json:"streams"`
}

type StreamST struct {
	URL          string `json:"url"`
	Status       bool   `json:"status"`
	OnDemand     bool   `json:"on_demand"`
	FPS          int    `json:"fps"`
	SamplingRate int    `json:"samplingRate"`
}

// 本质还不就是把配置文件从磁盘加载到内存里
func loadConfig() *configST {
	var tmp configST
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		log.Fatalln(err)
	}
	return &tmp
}
