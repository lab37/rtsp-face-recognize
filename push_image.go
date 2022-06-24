package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func monitAndPutNewImgToChan(fileName string, imgQueue chan image.Image) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	imgBytes, err := os.ReadFile(fileName)
	for {
		_, err := os.ReadFile(fileName)
		if err == nil {
			break
		}
		log.Println("waiting imgFile generate.......")
		time.Sleep(time.Duration(2) * time.Second)
	}
	err = watcher.Add(fileName)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("FileMonitor err", err)
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				imgBytes, _ = os.ReadFile(fileName)
				img, _ := jpeg.Decode(bytes.NewReader(imgBytes))
				if len(imgQueue) < cap(imgQueue) {
					imgQueue <- img
				} else {
					log.Println("img pipe is full")
				}

			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}
