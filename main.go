package main

import (
	"encoding/json"
	"image"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 建立人名传递通道
	nameQueue := make(chan string, 20)

	// 建立人名传递通道
	imgQueue := make(chan image.Image, 25)

	// 生成mqtt客户端，并连接mqtt服务器
	mqttClient := createMQTTClient(Config.MQTTserver, "faceRec-camera", Config.MQTTuserName, Config.MQTTpassword)
	defer mqttClient.Terminate()

	// 为了实时性这里采用监听绿米网关组播消息的方式来触发动作
	multicastMessageCh := make(chan message, 100)
	go udpMulticastReceiver("224.0.0.50", 9898, "", multicastMessageCh)
	go func() {
		for {
			select {
			case message := <-multicastMessageCh:
				multicastData := dataST{}
				multicastPayload := payload{}
				json.Unmarshal(message.Data, &multicastData)
				json.Unmarshal([]byte(multicastData.Payload), &multicastPayload)
				if multicastPayload.RGB > 0 {
					log.Println("收到人员活动报告..............................")
					// 推送MQTT消息给hass, 由hass显示门口画面。
					publishMQTTtopic(mqttClient, `security\gate\motion`, "", 0)
					// 启动ffmpeg程序截图交给人脸识别程序来识别
					cmd := exec.Command("sh", "-c", Config.FFmpegScriptFile)
					cmd.Run()
				}
			}
		}
	}()

	// 生成人脸检测器
	faceDetectClassifier := getFaceDetectClassifier(`cascade/facefinder`)

	// 加载已知人脸数据库
	faceDescriptions, names := loadFacesDatabase(`face-data.json`)

	// 生成对应模型的人脸识别器
	faceRecogizer := getFaceRecognizer("testdata", faceDescriptions)
	defer faceRecogizer.Close()

	// 开协程收集图片
	go monitAndPutNewImgToChan(Config.ImgFileName, imgQueue)

	log.Println("开启图片识别协程..............................")
	// 开协识别图片
	go func() {
		for {
			select {
			case tmpImg := <-imgQueue:
				if len(imgQueue) > 10 {
					log.Println("图片积压！, 当前已积压：", len(imgQueue), "张图片。你能换个CPU吗!")
				}
				numberOfFace := detectFace(faceDetectClassifier, tmpImg)
				if numberOfFace > 0 {
					recognizeFaceAndPushName(faceRecogizer, names, tmpImg, nameQueue)
				}

			}
		}

	}()

	log.Println("开启统计人名协程..............................")
	// 开协程每5秒统计一下来客, 并进行人脸播报
	go func() {

		mqttTicker := time.NewTicker(time.Second * 3)
		defer mqttTicker.Stop()
		// 建立人名统计映射
		nameCount := make(map[string]int)
		for {
			select {
			case cName := <-nameQueue:
				nameCount[cName] = nameCount[cName] + 1
			case <-mqttTicker.C:
				message := ""
				nums := 0
				cAnonymousNum := nameCount["anonymous"]
				for key, value := range nameCount {
					if value > 1 && key != "anonymous" {
						nums = nums + 1
						message = message + key + ","
					}
					nameCount[key] = 0
				}
				switch {
				case nums == 0:
					if cAnonymousNum > 3 {
						message = message + "有陌生人来了"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					}
				case nums > 0:
					if cAnonymousNum > 3 {
						message = message + "来了, 带着陌生人"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					} else {
						message = message + "来了"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					}
				}
			}

		}
	}()
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println(sig)
		done <- true
	}()
	log.Println("Server Start Awaiting Signal")
	<-done
	log.Println("Exiting")
}
