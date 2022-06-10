package main

import (
	"image"
	"log"
	"strings"
	"time"

	"github.com/nfnt/resize"

	"github.com/Kagami/go-face"
)

func main() {
	imgStream := make(chan image.Image, 60)

	rtspClient, frameDecoderSingle := getRtspFrameDecoder(config.Streams["xiaoyi"].URL)
	defer rtspClient.Close()

	faceDetectClassifier := getFaceDetectClassifier(`cascade/facefinder`)

	// 把通道当队列用，也是没谁了。当然，不要在意这些细节，它能工作。
	// Because I don't know other good method, so I used this stupid one.
	go putImagesToChan(rtspClient, frameDecoderSingle, imgStream)

	faceDescriptions, names := loadFacesDatabase(`face-data.json`)

	faceRecogizer := getFaceRecognizer("testdata", faceDescriptions)
	defer faceRecogizer.Close()

	// 朝巴电脑性能不行，无法对每张图片进行检测，只能以设置定时器的方式平均采样。
	// Bad computer, be coerced.
	ticker := time.NewTicker(time.Millisecond * time.Duration(config.Streams["xiaoyi"].SamplingRate))
	defer ticker.Stop()

	for {
		select {
		case tmpImg := <-imgStream:
			if len(imgStream) > 10 {
				log.Println("图片积压！, 当前已积压：", len(imgStream), "张图片。你能换个CPU吗!")
			}
			select { //平均采样图片进行检测. Average in time.
			case <-ticker.C:
				smallImg := resize.Resize(320, 0, tmpImg, resize.Lanczos3)
				numberOfFace := detectFace(faceDetectClassifier, smallImg)
				if numberOfFace > 0 {
					recImg := resize.Resize(720, 0, tmpImg, resize.Lanczos3)
					// 我那朝巴电脑对这个不敏感，我特么也不知道这里应不应该开协程。
					// Bad computer, it didn't tell me anything useful.
					go recognizeFaceAndDoSomething(faceRecogizer, names, recImg)
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
func recognizeFaceAndDoSomething(faceRecogizer *face.Recognizer, names []string, tmpImg image.Image) {

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
		log.Println("来了", results.anonymousNum+len(results.names), "个人:", strings.Join(results.names, ", "))
		return
	}
	if len(results.names) == 0 {
		log.Println("来了", results.anonymousNum, "个陌生人")
		return
	}

	log.Println("来了", results.anonymousNum+len(results.names), "个人:", strings.Join(results.names, ", "), "还有", results.anonymousNum, "陌生人")
}
