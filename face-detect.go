package main

import (
	"image"
	"io/ioutil"
	"log"

	pigo "github.com/esimov/pigo/core"
)

// 使用特定的模型对象来检测图片中的人脸。计算出人脸在图片的位置与置信度，返回人脸个数。
func detectFace(classifier *pigo.Pigo, img image.Image) (numberOfFace int) {

	angle := 0.0 // cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians

	src := pigo.ImgToNRGBA(img)

	pixels := pigo.RgbToGrayscale(src)
	cols, rows := src.Bounds().Max.X, src.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}
	// 下面的注释是原厂带的，不要问我为什么没有中文注释，因为我英语暂时还没过四级。
	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)
	numberOfFace = len(dets)
	if numberOfFace > 0 {
		// log.Println("来了，老弟， 好嗨哟！")
	}
	return
}

// 解析模型文件，返回解析后得到的模型对象。
func getFaceDetectClassifier(modelPath string) (classifier *pigo.Pigo) {
	cascadeFile, err := ioutil.ReadFile(modelPath)
	if err != nil {
		log.Fatalln("什么狗FaceDetect模型路径, 读不了: ", err)
	}

	p := pigo.NewPigo()

	// 下面的注释也是原厂带的，但这个我好像认识不少单词，只是整句话的意思不认识。
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err = p.Unpack(cascadeFile)
	if err != nil {
		log.Fatalln("路径倒是没问题了，但这熊文件解析不了啊: ", err)
	}
	return
}
