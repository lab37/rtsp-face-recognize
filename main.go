package main

import (
	"image"
	"log"
	"time"

	"github.com/nfnt/resize"
)

func main() {
	imgStream := make(chan image.Image, 60)

	rtspClient, frameDecoderSingle := getRtspFrameDecoder(config.Streams["xiaoyi"].URL)
	defer rtspClient.Close()

	faceDetectClassifier := getFaceDetectClassifier(`cascade/facefinder`)

	faceDescriptions, names := loadFacesDatabase(`face-data.json`)

	faceRecogizer := getFaceRecognizer("testdata", faceDescriptions)
	defer faceRecogizer.Close()

	// 把通道当队列用，也是没谁了。当然，不要在意这些细节，它能工作。
	// Because I don't know other good method, so I used this stupid one.
	go putImagesToChan(rtspClient, frameDecoderSingle, imgStream)

	// 朝巴电脑性能不行，无法对每张图片进行检测，只能以设置定时器的方式平均采样。
	// Bad computer, be coerced.
	ticker := time.NewTicker(time.Millisecond * time.Duration(config.Streams["xiaoyi"].SamplingRate))
	defer ticker.Stop()

	for {
		select {
		case tmpImg := <-imgStream:
			if len(imgStream) > 10 {
				log.Println("图片积压！, 当前已积压：", len(imgStream), "张图片。你能换个CPU吗！")
			}
			select { //平均采样图片进行检测. Average in time.
			case <-ticker.C:
				smallImg := resize.Resize(320, 0, tmpImg, resize.Lanczos3)
				numberOfFace := detectFace(faceDetectClassifier, smallImg)
				if numberOfFace > 0 {
					// 我那朝巴电脑对这个不敏感，我特么也不知道这里应不应该开协程。
					// Bad computer, it didn't tell me anything useful.
					go faceRec(faceRecogizer, tmpImg, names)
				}

			default:
				//log.Println("时间不到，不上班。Ignore this picture，Go back to sleep")
			}
		default:
			//log.Println("这次不赖我，库里没货了. Not my fault, Stupid NIC")
		}

	}

}
