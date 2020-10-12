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
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	fps = 4
	outputDir = "./dist"
	)

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
	svc, err := ffmpeg.New(ff)
	_, err = ffmpeg.New(ff)
	if err != nil {
		fmt.Println("new ffmpeg")
		fmt.Println(err)
		return
	}

	//sourceFile := "./source/IMG_1532.MOV"

	sourceFile := "./source/LAE.MOV"
	fmt.Printf("Reading video file: %s\n", sourceFile)
	err = svc.ExtractFrames(sourceFile, outputDir, fps)
	fmt.Println("Finished creating frames from video")
	if err != nil {
		fmt.Println("failed extracting frames")
		fmt.Println(err)
	}

	// would be nice to stream the results instead of writing to disk

	var impedanceValues []float64
	var tempValues []float64
	var secondValues []float64
	var powerValues []int
	totalFiles := numFilesInDir(outputDir, "frames")
	fmt.Printf("found %d files\n", totalFiles)
	for i := 1; i < totalFiles; i++ {
		file := fmt.Sprintf("%s/frames%05d.jpg", outputDir, i)
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

		if len(power) > 0 && len(temp) > 0 && len(impedance) > 0 && len(seconds) > 0 {
			_, err := strconv.Atoi(power)
			if err != nil {
				printAllValues(temp, impedance, power, seconds, "power")
				continue
			}

			impedanceInt, err := strconv.Atoi(impedance)
			if err != nil {
				printAllValues(temp, impedance, power, seconds, "impedance")
				continue
			}

			tempInt, err := strconv.Atoi(temp)
			if err != nil {
				printAllValues(temp, impedance, power, seconds, "temp")
				continue
			}

			secondInt, err := strconv.Atoi(seconds)
			if err != nil {
				printAllValues(temp, impedance, power, seconds, "time")
				continue
			}

			powerInt, err := strconv.Atoi(power)
			if err != nil {
				printAllValues(temp, impedance, power, seconds, "power")
				continue
			}

			impedanceValues = append(impedanceValues, float64(impedanceInt))
			tempValues = append(tempValues, float64(tempInt))
			powerValues = append(powerValues, powerInt)
			if len(secondValues) > 0 {
				previousSecondsValue := secondValues[len(secondValues)-1]
				//fmt.Printf("-----\nprevious second: %f\n", previousSecondsValue)
				//fmt.Printf("current second: %d\n", secondInt)
				if int(math.Floor(previousSecondsValue)) == secondInt {
					newSecondsValue := previousSecondsValue + float64(1)/float64(fps)
					//newSecondsValue := previousSecondsValue + 0.25
					//fmt.Print(fmt.Sprintf("newSecondsValues: %f\n", newSecondsValue))
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

	//tempRange := chart.Range()

	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name: "Time",
		},
		YAxis: chart.YAxis{
			Name: "Temperature",
			//Range: tempRange,
		},
		YAxisSecondary: chart.YAxis{
			Name: "Impedance",
		},
		Series: []chart.Series{tempSeries, impedanceSeries},
	}

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
	if err != nil {
		log.Println(err)
	}

	writeToCSV(impedanceValues, tempValues, powerValues, secondValues)
}

func writeImage(num int, img image.Image, pt image.Point) {
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  120,
		Height: 100,
		Anchor:  pt,
		Mode:   cutter.TopLeft,
	})
	fOut, err := os.Create(fmt.Sprintf("%s/out_image%d.jpg", outputDir, num))
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
	client.SetImage(fmt.Sprintf("%s/out_image%d.jpg", outputDir, num))
	text, _ := client.Text()
	return text, nil
}

func printAllValues(temp string, impedance string, power string, seconds string, value string) {
	fmt.Printf("invalid value for %s\n  temp   : %s\n  impedance: %s\n  power    : %s\n seconds  : %s\n", value, temp, impedance, power, seconds)
}

func writeToCSV(impedances []float64, temps []float64, powers []int, seconds []float64){
	outputCSVFile := "result.csv"
	fmt.Printf("Writing to %s\n", outputCSVFile)
	file, err := os.Create(outputCSVFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var data = []string{"time", "temperature", "power", "impedance"}
	err = writer.Write(data)
	if err != nil {
		panic(err)
	}

	for i, second := range seconds {
		secondStr := fmt.Sprintf("%.2f", second)
		tempStr := fmt.Sprintf("%d", int(temps[i]))
		powerStr := fmt.Sprintf("%d", powers[i])
		impedanceStr := fmt.Sprintf("%d", int(impedances[i]))
		data = []string{secondStr, tempStr, powerStr, impedanceStr}
		err = writer.Write(data)
		if err != nil {
			panic(err)
		}
	}
}

func numFilesInDir(dir string, substr string) int {
	fileCount := 0
	// im sure theres a better way to do this
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), substr) {
			fileCount++
		}
		return nil
	})
	return fileCount
}