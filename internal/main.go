package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/lijo-jose/gffmpeg/pkg/gffmpeg"
	"github.com/lijo-jose/goutils/pkg/ffmpeg"
	"github.com/oliamb/cutter"
	"github.com/otiai10/gosseract"
	"github.com/wcharczuk/go-chart"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
)

const (fps = 4)

func main(){

	file, err := os.Create("result.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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
	//err = svc.ExtractFrames(sourceFile, "./dist", fps)
	if err != nil {
		fmt.Println("failed extracting frames")
		fmt.Println(err)
	}

	// would be nice to stream the results instead of writing to disk

	var impedanceValues []float64
	var tempValues []float64
	var secondValues []float64
	// 00001 -> 00044
	//totalFiles := 45
	totalFiles := 174
	for i := 1; i < totalFiles; i++ {
	//for i := 1; i < 15; i++ {
		file := fmt.Sprintf("./dist/frames%05d.jpg", i)
		fmt.Println(file)
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		img, _, err := image.Decode(f)
		f.Close()
		img = imaging.Rotate270(img)
		writeImage(1, img, image.Point{720, 220})
		writeImage(2, img, image.Point{975, 220})
		writeImage(3, img, image.Point{1220, 220})
		writeImage(4, img, image.Point{1255, 320})
		power, err := getTextFromImage(1)
		temp, err := getTextFromImage(2)
		impedance, err := getTextFromImage(3)
		seconds, err := getTextFromImage(4)

		//secondsNum, err := strconv.Atoi(seconds)
		//fmt.Println("checking condition")
		//fmt.Println(power)
		//fmt.Println(temp)
		//fmt.Println(impedance)
		//fmt.Println(seconds)
		if len(power) > 0 && len(temp) > 0 && len(impedance) > 0 && len(seconds) > 0 {
			fmt.Println(fmt.Sprintf("file      : %s", file))
			fmt.Println(fmt.Sprintf("power     : %s", power))
			fmt.Println(fmt.Sprintf("temp      : %s", temp))
			fmt.Println(fmt.Sprintf("impedance : %s", impedance))
			fmt.Println(fmt.Sprintf("seconds   : %s", seconds))
			_, err := strconv.Atoi(power)
			if err != nil {
				//panic(err)
				continue
			}

			impedanceInt, err := strconv.Atoi(impedance)
			if err != nil {
				//panic(err)
				continue
			}

			tempInt, err := strconv.Atoi(temp)
			if err != nil {
				//panic(err)
				continue
			}

			secondInt, err := strconv.Atoi(seconds)
			if err != nil {
				//panic(err)
				continue
			}

			//fmt.Println(impedance)
			//fmt.Println(impedanceInt)
			//fmt.Println(float64(impedanceInt))
			impedanceValues = append(impedanceValues, float64(impedanceInt))
			tempValues = append(tempValues, float64(tempInt))
			// if secondInt == the previous value then
			// secondInt = previous value + .25
			if len(secondValues) > 0 {
				previousSecondsValue := secondValues[len(secondValues)-1]
				if previousSecondsValue <= float64(secondInt) {
					//newSecondsValue := previousSecondsValue + float64(1/(fps+1))
					newSecondsValue := previousSecondsValue + 0.20
					fmt.Print(fmt.Sprintf("newSecondsValues: %f\n", newSecondsValue))
					secondValues = append(secondValues, newSecondsValue) // use FPS value
				} else {
					secondValues = append(secondValues, float64(secondInt))
				}
			} else {
				secondValues = append(secondValues, float64(secondInt))
			}
		}
	}
	impedanceSeries := chart.ContinuousSeries{
		XValues: secondValues,
		YValues: impedanceValues,
		YAxis: chart.YAxisType(1),
	}
	tempSeries := chart.ContinuousSeries{
		XValues: secondValues,
		YValues: tempValues,
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Time",
		},
		YAxis: chart.YAxis{
			Name: "Temperature",
		},
		YAxisSecondary: chart.YAxis{
			Name: "Impedance",
		},
		Series: []chart.Series{tempSeries, impedanceSeries},
	}

	fmt.Println(impedanceValues)
	fmt.Println(tempValues)
	fmt.Println(secondValues)

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println(buffer.Len())
	im, _, err := image.Decode(bytes.NewReader(buffer.Bytes()))
	out, _ := os.Create("./img.jpg")
	defer out.Close()

	var opts jpeg.Options
	opts.Quality = 100

	err = jpeg.Encode(out, im, &opts)
	//jpeg.Encode(out, img, nil)
	if err != nil {
		log.Println(err)
	}
}

func writeImage(num int, img image.Image, pt image.Point) {
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  120,
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

func getTextFromImage(num int) (string, error) {
	client := gosseract.NewClient()
	client.SetWhitelist("0123456789W")
	defer client.Close()
	client.SetImage(fmt.Sprintf("./dist/out_image%d.jpg", num))
	text, _ := client.Text()
	//return strconv.Atoi(text)
	return text, nil
}
