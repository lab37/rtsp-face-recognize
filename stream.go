package main

import (
	"log"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/cgo/ffmpeg"

	"image"

	"github.com/deepch/vdk/format/rtspv2"
)

func putImagesToChan(url string, imgStream chan image.Image) {
	rtspClient, frameDecoderSingle := getRtspFrameDecoder(url)
	defer rtspClient.Close()
	videoTestTime := time.NewTimer(20 * time.Second)
	for {
		select {
		case <-videoTestTime.C:
			log.Println("可能卡了, 已经20秒没有收到图像了")
		case signals := <-rtspClient.Signals:
			if signals == rtspv2.SignalStreamRTPStop {
				log.Println("RTSP流停止了, 正在重新连接")
				rtspClient, frameDecoderSingle = getRtspFrameDecoder(url)
			}
		// 有时处理速度不行，还是会导致RTSPClient.OutgoingPacketQueue通道中的数据包积压。
		// RTSPClient.OutgoingPacketQueue的默认长度是3000，但积到2000个包的时候它自己会提示并丢包。
		case packetAV := <-rtspClient.OutgoingPacketQueue:
			if packetAV.IsKeyFrame {
				videoTestTime.Reset(20 * time.Second)
			}

			// 数据包积压超过24个将打印提示。
			if len(rtspClient.OutgoingPacketQueue) > 24 {
				log.Println("网络视频流数据包积压,当前已积压", len(rtspClient.OutgoingPacketQueue), "个数据包, 这都能忍?你真不打算换个CPU么.")
			}

			// 不管用不用这一帧都必须解码，解码函数中有一部分功能是清空packetAV结构。
			// packetAV结构不清空无法复用，后面的数据包会发生错误。水平有限无法深入Cgo代码找清空逻辑。
			if pic, err := frameDecoderSingle.DecodeSingle(packetAV.Data); err == nil && pic != nil {
				select {
				case imgStream <- &pic.Image:
				default:
				}

			}

		}

	}
}

//连接RTSP地址, 返回客户端结构和解码器结构.
// connect rtsp camera, convert  video stream to images.
func getRtspFrameDecoder(url string) (rtspClient *rtspv2.RTSPClient, frameDecoderSingle *ffmpeg.VideoDecoder) {

	rtspClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: url, DisableAudio: true, DialTimeout: 3 * time.Second, ReadWriteTimeout: 3 * time.Second, Debug: false})

	if err != nil {
		log.Fatalln(err)
	}

	audioOnly := true

	var videoIDX int
	for i, codec := range rtspClient.CodecData {
		if codec.Type().IsVideo() {
			audioOnly = false
			videoIDX = i
		}
	}

	if !audioOnly {
		frameDecoderSingle, err = ffmpeg.NewVideoDecoder(rtspClient.CodecData[videoIDX].(av.VideoCodecData))
		if err != nil {
			log.Fatalln(err)
			return
		}
	}

	return
}
