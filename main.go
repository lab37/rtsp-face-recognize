package main

import (
	"log"
	"time"

	"github.com/nfnt/resize"

	"os"
	"os/signal"
	"syscall"
)

func main() {
	go serveHTTP()
	go serveStreams()
	// 建立人名传递通道
	nameQueue := make(chan string, 20)

	// 建立MQTT连接
	mqttClient := createMQTTClient(Config.Server.MQTTserver, "faceRec-camera", Config.Server.MQTTuserName, Config.Server.MQTTpassword)
	defer mqttClient.Terminate()

	// 生成人脸检测器
	faceDetectClassifier := getFaceDetectClassifier(`cascade/facefinder`)

	// 加载已知人脸数据库
	faceDescriptions, names := loadFacesDatabase(`face-data.json`)

	// 生成对应模型的人脸识别器
	faceRecogizer := getFaceRecognizer("testdata", faceDescriptions)
	defer faceRecogizer.Close()

	// 采样图片进行检测，考虑到性能问题只能以设置定时器的方式平均采样。

	for k, v := range Config.Streams {
		go func(streamName string, stream StreamST) {
			snapshotTicker := time.NewTicker(time.Millisecond * time.Duration(200))
			defer snapshotTicker.Stop()
			for {
				select {
				case <-snapshotTicker.C:
					if len(stream.ImgQueue) > 10 {
						log.Println(streamName, "流中的图片积压！, 当前已积压：", len(stream.ImgQueue), "张图片。你能换个CPU吗!")
					}
					tmpImg := <-stream.ImgQueue
					go func() {
						smallImg := resize.Resize(320, 0, tmpImg, resize.Lanczos3)
						numberOfFace := detectFace(faceDetectClassifier, smallImg)
						if numberOfFace > 0 {
							recImg := resize.Resize(720, 0, tmpImg, resize.Lanczos3)
							recognizeFaceAndPushName(faceRecogizer, names, recImg, nameQueue)
						}

					}()
				default:
					<-stream.ImgQueue
					//	log.Println("作废一张图片")
				}
			}
		}(k, v)
	}

	// 开协程每5秒统计一下来客, 并进行人脸播报
	go func() {

		mqttTicker := time.NewTicker(time.Second * 5)
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
					if cAnonymousNum > 5 {
						message = message + "有陌生人来了"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					}
				case nums > 0:
					if cAnonymousNum > 5 {
						message = message + "来了, 带着陌生人"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					} else {
						message = message + "来了"
						log.Println(message)
						publishMQTTtopic(mqttClient, `homeassistant\camera\facerec`, message, 0)
					}

				default:

				}

			default:
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
