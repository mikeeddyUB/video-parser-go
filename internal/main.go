package main

import (
	"fmt"
	"github.com/lijo-jose/gffmpeg/pkg/gffmpeg"
	"github.com/lijo-jose/goutils/pkg/ffmpeg"
	"github.com/otiai10/gosseract"
	"github.com/oliamb/cutter"
	"github.com/disintegration/imaging"
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

	//err = svc.ExtractFrames("./IMG_1532.MOV", "./", 1)
	if err != nil {
		fmt.Println("failed extracting frames")
		fmt.Println(err)
	}

	// would be nice to stream the results instead of writing to disk

	// now extract the times
	f, err := os.Open("frames00017.jpg")
	defer f.Close()
	img, _, err := image.Decode(f)
	croppedImg, err := cutter.Crop(img, cutter.Config{
		//Width:  300,
		//Height: 100,
		//Anchor: image.Point{580, 340},
		//Mode:   cutter.TopLeft, // optional, default value
		Width:  80,
		Height: 60,
		Anchor: image.Point{585, 345},
		Mode:   cutter.TopLeft, // optional, default value
	})

	sharpenedImage := imaging.Sharpen(croppedImg, 1.5)
	dstImage := imaging.AdjustContrast(sharpenedImage, 15)

	fOut, err := os.Create("out_image.jpg")
	if err != nil {
		panic(err)
	}
	defer fOut.Close()

	opt := jpeg.Options{
		Quality: 100, // 0-100
	}
	err = jpeg.Encode(fOut, dstImage, &opt)
	if err != nil {
		panic(err)
	}


	// the images need to be cropped
	client := gosseract.NewClient()
	//client.SetConfigFile()
	//client.SetWhitelist("0123456789")
	defer client.Close()
	client.SetImage("out_image.jpg")
	text, _ := client.Text()
	fmt.Println(text)
}