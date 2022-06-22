package main

import (
	"errors"
	"image"
	"log"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/cgo/ffmpeg"
	"github.com/deepch/vdk/format/rtspv2"
)

var (
	ErrorStreamExitNoVideoOnStream = errors.New("Stream Exit No Video On Stream")
	ErrorStreamExitRtspDisconnect  = errors.New("Stream Exit Rtsp Disconnect")
	ErrorStreamExitNoViewer        = errors.New("Stream Exit On Demand No Viewer")
)

// 为每个配置的RTSP流开启一个协程建立连接, 并进行连接监测管理
func serveStreams() {
	for k, v := range Config.Streams {
		if !v.OnDemand {
			go RTSPWorkerSupervisor(k, v.URL, v.OnDemand, v.DisableAudio, v.Debug, v.VideoPacketQueue, v.ImgQueue)
		}
	}
}

//用于监测RTSP连接, 断线后1秒重连
func RTSPWorkerSupervisor(name, url string, OnDemand, DisableAudio, Debug bool, videoPacketQueue chan []byte, imgQueue chan image.Image) {
	defer Config.RunUnlock(name)
	for {
		log.Println("Stream Try Connect", name)
		err := RTSPWorker(name, url, OnDemand, DisableAudio, Debug, videoPacketQueue, imgQueue)
		if err != nil {
			log.Println(err)
			Config.LastError = err
		}
		if OnDemand && !Config.HasViewer(name) {
			log.Println(ErrorStreamExitNoViewer)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// RTSP连接建立,并获取数据包
func RTSPWorker(name, url string, OnDemand, DisableAudio, Debug bool, videoPacketQueue chan []byte, imgQueue chan image.Image) error {
	keyTest := time.NewTimer(20 * time.Second)
	clientTest := time.NewTimer(20 * time.Second)
	//add next TimeOut
	RTSPClient, err := rtspv2.Dial(rtspv2.RTSPClientOptions{URL: url, DisableAudio: DisableAudio, DialTimeout: 3 * time.Second, ReadWriteTimeout: 3 * time.Second, Debug: Debug})
	if err != nil {
		return err
	}
	defer RTSPClient.Close()
	log.Println("返回的所有流的配置信息：", RTSPClient.CodecData)
	if RTSPClient.CodecData != nil {
		Config.coAd(name, RTSPClient.CodecData)
	}
	var AudioOnly bool
	if len(RTSPClient.CodecData) == 1 && RTSPClient.CodecData[0].Type().IsAudio() {
		AudioOnly = true
	}
	if !AudioOnly {
		log.Println("开启协程往照片队列中放入照片..........")
		go putImagesToQueue(name, videoPacketQueue, imgQueue)
	}

	for {
		select {
		case <-clientTest.C:
			if OnDemand {
				if !Config.HasViewer(name) {
					return ErrorStreamExitNoViewer
				} else {
					clientTest.Reset(20 * time.Second)
				}
			}
		case <-keyTest.C:
			return ErrorStreamExitNoVideoOnStream
		case signals := <-RTSPClient.Signals:
			switch signals {
			case rtspv2.SignalCodecUpdate:
				log.Println("来了个信号codecUpdate")
				Config.coAd(name, RTSPClient.CodecData)
			case rtspv2.SignalStreamRTPStop:
				log.Println("来了个信号Disconnect")
				return ErrorStreamExitRtspDisconnect

			}
		case packetAV := <-RTSPClient.OutgoingPacketQueue:
			if AudioOnly || packetAV.IsKeyFrame {
				//	log.Println("来了个关键数据包")
				keyTest.Reset(20 * time.Second)
			}
			b := make([]byte, len(packetAV.Data))
			copy(b, packetAV.Data)
			select {
			case videoPacketQueue <- b:
				//log.Println("往视频帧队列中放入了一个包,  当前队列长度: ", len(videoPacketQueue))
			default:
				log.Println("视频帧数据包队列满了,  目前长度: ", len(videoPacketQueue))
			}

			Config.cast(name, *packetAV)
		}
	}
}

// 解码数据包, 并把解码得到的图片放到队列imgStream中。
func putImagesToQueue(name string, videoPacketQueue chan []byte, imgQueue chan image.Image) {
	audioOnly := true
	var videoIDX int
	for i, codec := range Config.Streams[name].Codecs {
		if codec.Type().IsVideo() {
			audioOnly = false
			videoIDX = i
		}
	}
	if audioOnly {
		log.Println("rtsp流中没有视频, 只有音频. putImagesToQueue协程退出")
		return
	}
	log.Println("生成流", name, "的照片解码器")
	frameDecoderSingle, err := ffmpeg.NewVideoDecoder(Config.Streams[name].Codecs[videoIDX].(av.VideoCodecData))
	if err != nil {
		log.Fatalln("生成流", name, "的照片解码器时发生错误：", err)
		return
	}
	for {
		select {
		case packetAvDate := <-Config.Streams[name].VideoPacketQueue:
			if pic, err := frameDecoderSingle.DecodeSingle(packetAvDate); err == nil && pic != nil {
				if len(imgQueue) < cap(imgQueue) {
					imgQueue <- &pic.Image
				} else {
					log.Println("图片队列已满, 无法放入新图片")
				}
			}
		default:
			//	log.Println("视频帧数据包队列中没有数据")
		}

	}

}
