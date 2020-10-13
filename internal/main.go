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
	"image"
	"image/jpeg"
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

	file, err := os.Create(fmt.Sprintf("%s/result.csv", outputDir))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	ff, err := gffmpeg.NewGFFmpeg("/usr/bin/ffmpeg")
	if err != nil {
		fmt.Println(err)
		return
	}
	svc, err := ffmpeg.New(ff)
	_, err = ffmpeg.New(ff)
	if err != nil {
		fmt.Println(err)
		return
	}

	sourceFile := "./source/LAE.MOV"
	fmt.Printf("Reading video file: %s\n", sourceFile)
	err = svc.ExtractFrames(sourceFile, outputDir, fps)
	fmt.Println("Finished creating frames from video")
	if err != nil {
		fmt.Println("failed extracting frames")
		fmt.Println(err)
	}

	var impedanceValues []float64
	var tempValues []float64
	var secondValues []float64
	var powerValues []int

	totalFiles := numFilesInDir(outputDir, "frames")
	fmt.Printf("Found %d files\n", totalFiles)
	client := gosseract.NewClient()
	client.SetWhitelist("0123456789F")
	defer client.Close()
	for i := 1; i < totalFiles; i++ {
		impedanceValues, tempValues, secondValues, powerValues = loopStuff(i, client)
	}

	writeToCSV(impedanceValues, tempValues, powerValues, secondValues)
}

func loopStuff(i int, client *gosseract.Client) ([]float64, []float64, []float64, []int){
	var impedanceValues []float64
	var tempValues []float64
	var secondValues []float64
	var powerValues []int

	file := fmt.Sprintf("%s/frames%05d.jpg", outputDir, i)
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(f)
	defer f.Close()
	img = imaging.Rotate270(img)
	power, err := extractText(client, img, image.Point{720, 220})
	temp, err := extractText(client, img, image.Point{975, 220})
	impedance, err := extractText(client, img, image.Point{1220, 220})
	seconds, err := extractText(client, img, image.Point{1255, 320})

	if len(power) > 0 && len(temp) > 0 && len(impedance) > 0 && len(seconds) > 0 {
		_, err := strconv.Atoi(power)
		if err != nil {
			printAllValues(temp, impedance, power, seconds, "power", file)
			return impedanceValues, tempValues, secondValues, powerValues
		}

		impedanceInt, err := strconv.Atoi(impedance)
		if err != nil {
			printAllValues(temp, impedance, power, seconds, "impedance", file)
			return impedanceValues, tempValues, secondValues, powerValues
		}

		tempInt, err := strconv.Atoi(temp)
		if err != nil {
			printAllValues(temp, impedance, power, seconds, "temp", file)
			return impedanceValues, tempValues, secondValues, powerValues
		}

		secondInt, err := strconv.Atoi(seconds)
		if err != nil {
			printAllValues(temp, impedance, power, seconds, "time", file)
			return impedanceValues, tempValues, secondValues, powerValues
		}

		powerInt, err := strconv.Atoi(power)
		if err != nil {
			printAllValues(temp, impedance, power, seconds, "power", file)
			return impedanceValues, tempValues, secondValues, powerValues
		}

		impedanceValues = append(impedanceValues, float64(impedanceInt))
		tempValues = append(tempValues, float64(tempInt))
		powerValues = append(powerValues, powerInt)
		if len(secondValues) > 0 {
			previousSecondsValue := secondValues[len(secondValues)-1]
			if int(math.Floor(previousSecondsValue)) == secondInt {
				newSecondsValue := previousSecondsValue + float64(1)/float64(fps)
				secondValues = append(secondValues, newSecondsValue)
			} else {
				secondValues = append(secondValues, float64(secondInt))
			}
		} else {
			secondValues = append(secondValues, float64(secondInt))
		}
	}
	return impedanceValues, tempValues, secondValues, powerValues
}

func extractText(client *gosseract.Client, img image.Image, pt image.Point) (string, error) {
	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  120,
		Height: 100,
		Anchor:  pt,
		Mode:   cutter.TopLeft,
	})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, croppedImg, nil)
	if err != nil {
		return "", err
	}
	client.SetImageFromBytes(buf.Bytes())
	text, _ := client.Text()
	return text, nil
}

func printAllValues(temp string, impedance string, power string, seconds string, value string, file string) {
	fmt.Printf("invalid value for %s\n  temp   : %s\n  impedance: %s\n  power    : %s\n seconds  : %s\n file: %s\n", value, temp, impedance, power, seconds, file)
}

func writeToCSV(impedances []float64, temps []float64, powers []int, seconds []float64){
	outputCSVFile := fmt.Sprintf("%s/result.csv", outputDir)
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