package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/lijo-jose/gffmpeg/pkg/gffmpeg"
	"github.com/lijo-jose/goutils/pkg/ffmpeg"
	"github.com/oliamb/cutter"
	"github.com/otiai10/gosseract"
	"image"
	"image/jpeg"
	"os"
)


func main(){

	ff, err := gffmpeg.NewGFFmpeg("/usr/bin/ffmpeg")
	if err != nil {
		fmt.Println("calling ffmpeg")
		fmt.Println(err)
		return
	}
	//svc, err := ffmpeg.New(ff)
	_, err = ffmpeg.New(ff)
	if err != nil {
		fmt.Println("new ffmpeg")
		fmt.Println(err)
		return
	}

	//sourceFile := "./source/IMG_1532.MOV"
	//sourceFile := "./source/LAE.MOV"
	//err = svc.ExtractFrames(sourceFile, "./dist", 1)
	if err != nil {
		fmt.Println("failed extracting frames")
		fmt.Println(err)
	}

	// would be nice to stream the results instead of writing to disk

	// now extract the times
	f, err := os.Open("./dist/frames00017.jpg")
	defer f.Close()
	img, _, err := image.Decode(f)
	img = imaging.Rotate270(img)
	//img, err = cutter.Crop(img, cutter.Config{
	//	Width:  800,
	//	Height: 170,
	//	//Anchor: image.Point{580, 340},
	//	//Mode:   cutter.TopLeft, // optional, default value
	//	//Width:  80,
	//	//Height: 60,
	//	//Anchor: image.Point{585, 345},
	//	Anchor: image.Point{700, 220},
	//	Mode:   cutter.TopLeft, // optional, default value
	//})
	// 4 numbers we need to get

	//img = imaging.Sharpen(img, 1.5)
	//img = imaging.Invert(img)
	//dstImage := imaging.AdjustContrast(img, 15)

	writeImage(1, img, image.Point{700, 220})
	writeImage(2, img, image.Point{975, 220})
	writeImage(3, img, image.Point{1220, 220})
	writeImage(4, img, image.Point{1255, 320})
	fmt.Println(fmt.Sprintf("power     : %s", getTextFromImage(1)))
	fmt.Println(fmt.Sprintf("temp      : %s", getTextFromImage(2)))
	fmt.Println(fmt.Sprintf("impedence : %s", getTextFromImage(3)))
	fmt.Println(fmt.Sprintf("seconds   : %s", getTextFromImage(4)))
}

func writeImage(num int, img image.Image, pt image.Point) {
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  100,
		Height: 100,
		Anchor:  pt,
		Mode:   cutter.TopLeft,
	})
	fOut, err := os.Create(fmt.Sprintf("./dist/out_image%d.jpg", num))
	if err != nil {
		panic(err)
	}
	defer fOut.Close()

	opt := jpeg.Options{
		Quality: 100, // 0-100
	}
	err = jpeg.Encode(fOut, croppedImg, &opt)
	if err != nil {
		panic(err)
	}
}

func getTextFromImage(num int) string {
	client := gosseract.NewClient()
	client.SetWhitelist("0123456789W")
	defer client.Close()
	client.SetImage(fmt.Sprintf("./dist/out_image%d.jpg", num))
	text, _ := client.Text()
	return text
}