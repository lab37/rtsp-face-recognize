package main

import (
	"image"
	"log"
	"time"

	"github.com/nfnt/resize"

	"github.com/Kagami/go-face"
)

func main() {
	// 建立照片传递通道
	imgStream := make(chan image.Image, 60)

	// 建立人名传递通道
	nameStream := make(chan string, 20)

	// 建立MQTT连接
	mqttClient := createMQTTClient(config.Streams["xiaoyi"].MQTTserver, "faceRec-camera", config.Streams["xiaoyi"].MQTTuserName, config.Streams["xiaoyi"].MQTTpassword)
	defer mqttClient.Terminate()

	// 生成人脸检测器
	faceDetectClassifier := getFaceDetectClassifier(`cascade/facefinder`)

	// 把通道当队列用，也是没谁了。当然，不要在意这些细节，它能工作。
	// Because I don't know other good method, so I used this stupid one.
	go putImagesToChan(config.Streams["xiaoyi"].URL, imgStream)

	faceDescriptions, names := loadFacesDatabase(`face-data.json`)

	faceRecogizer := getFaceRecognizer("testdata", faceDescriptions)
	defer faceRecogizer.Close()

	mqttTicker := time.NewTicker(time.Second * 5)

	// 开协程每5秒统计一下来客
	go func() {
		// 建立人名统计映射
		nameCount := make(map[string]int)
		for {
			select {
			case cName := <-nameStream:
				nameCount[cName] = nameCount[cName] + 1
			case <-mqttTicker.C:
				message := ""
				nums := 0
				cAnonymousNum := nameCount["anonymous"]
				for key, value := range nameCount {
					if value > 3 && key != "anonymous" {
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

	// 朝巴电脑性能不行，无法对每张图片进行检测，只能以设置定时器的方式平均采样。
	// Bad computer, be coerced.
	snapshotTicker := time.NewTicker(time.Millisecond * time.Duration(config.Streams["xiaoyi"].SamplingRate))
	defer snapshotTicker.Stop()

	for {
		select {
		case tmpImg := <-imgStream:
			if len(imgStream) > 10 {
				log.Println("图片积压！, 当前已积压：", len(imgStream), "张图片。你能换个CPU吗!")
			}
			select { //平均采样图片进行检测. Average in time.
			case <-snapshotTicker.C:
				smallImg := resize.Resize(320, 0, tmpImg, resize.Lanczos3)
				numberOfFace := detectFace(faceDetectClassifier, smallImg)
				if numberOfFace > 0 {
					recImg := resize.Resize(720, 0, tmpImg, resize.Lanczos3)
					// 我那朝巴电脑对这个不敏感，我特么也不知道这里应不应该开协程。
					// Bad computer, it didn't tell me anything useful.
					go recognizeFaceAndPushName(faceRecogizer, names, recImg, nameStream)
				}

			default:
				//log.Println("时间不到，不上班。Ignore this picture，Go back to sleep")
			}
		default:
			//log.Println("这次不赖我，库里没货了. Not my fault, Stupid NIC")
		}

	}

}

// 这里是要交到协程里去执行的, 尽量不要有计算量太大的工作, 因为faceRec这个函数就已经很吃算力了。
func recognizeFaceAndPushName(faceRecogizer *face.Recognizer, names []string, tmpImg image.Image, nameStream chan string) {

	results, err := faceRec(faceRecogizer, tmpImg, names)

	if err != nil {
		//log.Println("图片都整不对, 你想让我干个啥, do个毛啊")
		return
	}

	if results.anonymousNum == 404 {
		//log.Println("图片上连个鸟儿都没有, 你想让我干个啥, do个毛啊")
		return
	}

	if results.anonymousNum == 0 {
		for _, name := range results.names {
			nameStream <- name
		}
		return
	}
	if len(results.names) == 0 {
		nameStream <- "anonymous"
		return
	}
	for _, name := range results.names {
		nameStream <- name
	}
	nameStream <- "anonymous"
}
