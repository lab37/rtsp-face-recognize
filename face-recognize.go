package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"io/ioutil"

	"log"
	"path/filepath"

	"github.com/Kagami/go-face"
)

type faceST struct {
	Name       string       `json:"name"`
	Descriptor [128]float32 `json:"descriptor"`
}

type recResult struct {
	anonymousNum int
	names        []string
}

// 人脸对比
// face-recognizor
func faceRec(rec *face.Recognizer, img image.Image, names []string) (results recResult, err error) {
	//start := time.Now()
	buf := new(bytes.Buffer)
	jpeg.Encode(buf, img, nil)
	rawImg := buf.Bytes()
	cFaces, err := rec.Recognize(rawImg)
	if err != nil {
		log.Println("要识别的图片有问题啊大哥", err)
		return
	}
	if len(cFaces) == 0 {
		// 用404代表没识别出东西来
		results.anonymousNum = 404
		return
	}

	numAnonymous := 0
	faceDescriptorIndex := 0
	for _, faceI := range cFaces {
		faceDescriptorIndex = rec.Classify(faceI.Descriptor)

		//为什么这么写, 原因是发现不认识的人经常被识别为第一个
		if faceDescriptorIndex < 1 {
			numAnonymous = numAnonymous + 1
			continue
		}
		results.names = append(results.names, names[faceDescriptorIndex])

	}
	results.anonymousNum = numAnonymous

	//elapsed := time.Since(start)
	//fmt.Println("该函数执行完成耗时：", elapsed)
	// 不要看我注释掉的这几行代码了，这是我用来测试我那电脑蠢到了什么程度的。
	return
}

// 使用特定的模型生成人脸识别器，并将已知人脸特征数据填充到此结构中。
// generate  Recognizer  by  models
func getFaceRecognizer(modelDir string, faceDescriptions []face.Descriptor) (rec *face.Recognizer) {
	modelsPath := filepath.Join(modelDir, "models")

	rec, err := face.NewRecognizer(modelsPath)
	if err != nil {
		log.Fatalln("人脸识别模型文件读不了,原因自己找去吧, 再见了宝贝儿:", err)
		return
	}
	var descriptorIndexs []int32

	// 提取识别的脸部特征与对应的索引i
	for i := 0; i < len(faceDescriptions); i++ {
		descriptorIndexs = append(descriptorIndexs, int32(i))
	}

	// Pass samples to the recognizer. 设置样本空间与对应的hash，当调用Classify鉴别时会返回对应的hash
	rec.SetSamples(faceDescriptions, descriptorIndexs)
	return
}

// 加载已有人脸数据，做为对比基准数据。人脸数据在face-data.json文件中。
// 别忘了再写一个用来从图片生成人脸特征数据的程序，不然这玩意儿怎么用。
// function name is the annotation.
func loadFacesDatabase(databasePath string) (faces []face.Descriptor, names []string) {
	var tmp []faceST
	data, err := ioutil.ReadFile(databasePath)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		log.Fatalln(err)
	}

	for _, f := range tmp {
		faces = append(faces, f.Descriptor)
		names = append(names, f.Name)
	}

	return
}

// 从图片中识别人脸并计算出人脸特征值，人脸识别顺序为从左到右。
// 这个函数废了，留着它只是为了用来万一图片上有多张脸的时候可以一下识别完。
// function name is the annotation.
func getFaceDescription(modelDir string, imagesPath string, numbers int) {
	modelsPath := filepath.Join(modelDir, "models")

	rec, err := face.NewRecognizer(modelsPath)
	if err != nil {
		log.Println("Can't init face recognizer:", err)
		return
	}
	defer rec.Close()
	// Recognize faces on that image. 返回的是一个[]Face（这是一个结构体切片,里面最重要的是Face.Descriptor）
	faces, err := rec.RecognizeFile(imagesPath)
	if err != nil {
		log.Println("Can't recognize:", err)
		return
	}
	if len(faces) != numbers {
		log.Println("Wrong number of faces")
		return
	}

	// 提取识别的脸部特征与对应的索引i
	for _, f := range faces {
		log.Println(f.Descriptor)
	}

}
