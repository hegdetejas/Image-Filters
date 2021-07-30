package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"math/rand"
	"os"
	"proj2/png"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SharedContext struct {
	mutex       *sync.Mutex
	cond        *sync.Cond
	completed   bool
	counter     int
	threadCount int
}

func main() {
	effects := make([]map[string]interface{}, 0)
	inputArgs := os.Args
	dataDir := inputArgs[1]

	effectsPathFile := "../data/effects.txt"
	effectsFile, err := os.Open(effectsPathFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(effectsFile)

	for scanner.Scan() {
		var effectStream map[string]interface{}
		if err := json.Unmarshal([]byte(scanner.Text()), &effectStream); err != nil {
			panic(err)
		}
		effects = append(effects, effectStream)
	}

	if len(inputArgs) == 2 {
		sequential(effects, dataDir)
	} else {
		paradigm := inputArgs[2]
		numThreads, _ := strconv.Atoi(inputArgs[3])

		if paradigm == "pipeline" {
			pipeline(effects, dataDir, numThreads)
		} else {
			bsp(effects, dataDir, numThreads)
		}
	}
}

func bsp(effects []map[string]interface{}, dataDir string, numThreads int) {
	sizeList := strings.Split(dataDir, "+")

	if len(sizeList) == 0 {
		sizeList = append(sizeList, dataDir)
	}

	for i := 0; i < len(sizeList); i++ {
		var mutex sync.Mutex
		condVar := sync.NewCond(&mutex)
		context := SharedContext{completed: false, threadCount: numThreads, cond: condVar, mutex: &mutex}
		workSchedule := make([]interface{}, 0)
		imageList := make([][]interface{}, 0)
		for j := 0; j < len(effects); j++ {
			sendList := make([]interface{}, 0)
			pngImg, err := png.Load(fmt.Sprintf("../data/in/%v/%v", sizeList[i], effects[j]["inPath"]))
			if err != nil {
				panic(err)
			}
			imgBounds := pngImg.In.Bounds()
			loopInterface := effects[j]["effects"].([]interface{})
			applyEffects := make([]string, len(loopInterface))

			for pos, value := range loopInterface {
				applyEffects[pos] = value.(string)
			}
			savePath := fmt.Sprintf("../data/out/%v_%v", sizeList[i], effects[j]["inPath"])
			sendList = append(sendList, pngImg, applyEffects, imgBounds, savePath)
			imageList = append(imageList, sendList)
		}
		for m := 0; m < len(imageList); m++ {
			xAmountStartRect := imageList[m][2].(image.Rectangle)
			xAmountWork := xAmountStartRect.Max.X / numThreads
			imageList[m] = append(imageList[m], xAmountWork)
		}
		workSchedule = append(workSchedule, imageList)
		workSchedule = append(workSchedule, 0)

		for k := 0; k < numThreads; k++ {
			go bspWorker(workSchedule, &context, k, numThreads)
		}

		for !context.completed {
		}

		for n := 0; n < len(imageList); n++ {
			img := imageList[n][0].(*png.Image)
			err := img.Save(imageList[n][3].(string))
			if err != nil {
				panic(err)
			}
		}
	}
}

func bspWorker(workSchedule []interface{}, ctx *SharedContext, multiplier int, numThreads int) {
	imageList := workSchedule[0].([][]interface{})

	for i := 0; i < len(imageList); i++ {
		var last bool
		effectsList := imageList[i][1].([]string)
		img := imageList[i][0].(*png.Image)
		for j := 0; j < len(effectsList); j++ {
			if j == len(effectsList)-1 {
				last = true
			}
			if effectsList[j] == "G" {
				img.Grayscale(last, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
			} else if effectsList[j] == "E" {
				img.EdgeDetection(last, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
			} else if effectsList[j] == "B" {
				img.Blur(last, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
			} else if effectsList[j] == "S" {
				img.Sharpen(last, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
			}
			if !last {
				copyImg(img.In, img.Temp, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
				copyImg(img.Out, img.In, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
				copyImg(img.Temp, img.Out, (multiplier * (imageList[i][4]).(int)), ((multiplier + 1) * (imageList[i][4]).(int)))
			}
		}
	}

	ctx.mutex.Lock()
	ctx.counter++
	if ctx.counter == ctx.threadCount {
		ctx.completed = true
	}
	if ctx.counter == ctx.threadCount {
		ctx.cond.Broadcast()
	} else {
		for ctx.counter != ctx.threadCount {
			ctx.cond.Wait()
		}
	}
	ctx.mutex.Unlock()
}

func copyImg(srcImg *image.RGBA64, dstImg *image.RGBA64, start int, end int) {
	srcImgBounds := srcImg.Bounds()

	yMin := srcImgBounds.Min.Y
	yMax := srcImgBounds.Max.Y

	for y := yMin; y < yMax; y++ {
		for x := start; x < end; x++ {
			pixel := srcImg.At(x, y)
			dstImg.Set(x, y, pixel)
		}
	}
}

func pipeline(effects []map[string]interface{}, dataDir string, numThreads int) {
	sizeList := strings.Split(dataDir, "+")
	if len(sizeList) == 0 {
		sizeList = append(sizeList, dataDir)
	}

	for i := 0; i < len(sizeList); i++ {
		imagesChan := filterService(effects, sizeList[i])
		done := make(chan bool, numThreads)
		for i := 0; i < numThreads; i++ {
			go worker(imagesChan, done, numThreads)
		}
		for j := 0; j < numThreads; j++ {
			<-done
		}
	}
}

func worker(imagesChan <-chan []interface{}, done chan bool, numThreads int) {
	for {
		requestedTask, more := <-imagesChan
		if more {
			var last bool
			requestedImg := requestedTask[0].(*png.Image)
			imgBounds := requestedImg.In.Bounds()

			xAmountStart := imgBounds.Min.X
			xWorkAmount := imgBounds.Max.X / numThreads
			yAmountStart := imgBounds.Min.Y
			yAmountEnd := imgBounds.Max.Y

			for i := 0; i < len(requestedTask[1].([]string)); i++ {
				if i == len(requestedTask[1].([]string))-1 {
					last = true
				}
				outImages := make(chan *image.RGBA64, numThreads)
				defer close(outImages)
				doneFilter := make(chan bool, numThreads)
				defer close(doneFilter)
				for j := 0; j < numThreads; j++ {
					if j == numThreads-1 {
						r := image.Rect(xAmountStart, yAmountStart, imgBounds.Max.X, yAmountEnd)
						wholeImg := requestedImg.In
						subImg := wholeImg.SubImage(r)
						inBounds := subImg.Bounds()
						inImgRGBA64 := image.NewRGBA64(image.Rect(0, 0, inBounds.Dx(), inBounds.Dy()))
						imagePointer := image.Point{xAmountStart, yAmountStart}
						draw.Draw(inImgRGBA64, inImgRGBA64.Bounds(), subImg, imagePointer, draw.Src)

						go applyFilter(doneFilter, requestedImg, requestedTask[1].([]string)[i], last, outImages)
					} else {
						r := image.Rect(xAmountStart, yAmountStart, xAmountStart+xWorkAmount, yAmountEnd)
						wholeImg := requestedImg.In
						subImg := wholeImg.SubImage(r)
						inBounds := subImg.Bounds()
						inImgRGBA64 := image.NewRGBA64(image.Rect(0, 0, inBounds.Dx(), inBounds.Dy()))
						imagePointer := image.Point{xAmountStart, yAmountStart}
						draw.Draw(inImgRGBA64, inImgRGBA64.Bounds(), subImg, imagePointer, draw.Src)

						go applyFilter(doneFilter, requestedImg, requestedTask[1].([]string)[i], last, outImages)
						xAmountStart += xWorkAmount
					}
				}
				for k := 0; k < numThreads; k++ {
					<-doneFilter
				}
				r := image.Rect(imgBounds.Min.X, imgBounds.Min.Y, imgBounds.Max.X, imgBounds.Max.Y)
				outImg := image.NewRGBA64(r)
				outBounds := outImg.Bounds()
				for m := 0; m < numThreads; m++ {
					inImgRGBA64 := <-outImages
					draw.Draw(outImg, outBounds, inImgRGBA64, inImgRGBA64.Bounds().Min, draw.Src)
				}
				if !last {
					copyImg(requestedImg.In, requestedImg.Temp, outBounds.Min.X, outBounds.Max.X)
					copyImg(requestedImg.Out, requestedImg.In, outBounds.Min.X, outBounds.Max.X)
					copyImg(requestedImg.Temp, requestedImg.Out, outBounds.Min.X, outBounds.Max.X)
				}
			}
			err := requestedImg.Save(fmt.Sprintf(requestedTask[2].(string)))
			if err != nil {
				panic(err)
			}
		} else {
			done <- true
			return
		}
	}
}

func applyFilter(doneFilter chan bool, img *png.Image, effect string, last bool, outImages chan *image.RGBA64) {
	if effect == "G" {
		img.Grayscale(last, -1, -1)
	} else if effect == "S" {
		img.Sharpen(last, -1, -1)
	} else if effect == "E" {
		img.EdgeDetection(last, -1, -1)
	} else if effect == "B" {
		img.Blur(last, -1, -1)
	}

	outImages <- img.Out
	doneFilter <- true
}

func filterService(effects []map[string]interface{}, dataDir string) <-chan []interface{} {
	channel := make(chan []interface{})

	go func() {
		defer close(channel)
		for i := 0; i < len(effects); i++ {
			channelList := make([]interface{}, 0)
			savePath := fmt.Sprintf("../data/out/%v_%v", dataDir, effects[i]["inPath"])
			pngImg, err := png.Load(fmt.Sprintf("../data/in/%v/%v", dataDir, effects[i]["inPath"]))

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			loopInterface := effects[i]["effects"].([]interface{})
			applyEffects := make([]string, len(loopInterface))

			for pos, value := range loopInterface {
				applyEffects[pos] = value.(string)
			}
			channelList = append(channelList, pngImg)
			channelList = append(channelList, applyEffects)
			channelList = append(channelList, savePath)

			channel <- channelList
			time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
		}
	}()
	return channel
}

func sequential(effects []map[string]interface{}, dataDir string) {
	sizeList := strings.Split(dataDir, "+")

	if len(sizeList) == 0 {
		sizeList = append(sizeList, dataDir)
	}

	for a := 0; a < len(sizeList); a++ {
		for i := 0; i < len(effects); i++ {
			pngImg, err := png.Load(fmt.Sprintf("../data/in/%v/%v", sizeList[a], effects[i]["inPath"]))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			loopInterface := effects[i]["effects"].([]interface{})
			applyEffects := make([]string, len(loopInterface))

			for pos, value := range loopInterface {
				applyEffects[pos] = value.(string)
			}
			for j := 0; j < len(applyEffects); j++ {
				if applyEffects[j] == "G" {
					if j == len(applyEffects)-1 {
						pngImg.Grayscale(true, -1, -1)
					} else {
						pngImg.Grayscale(false, -1, -1)
						pngImg.Temp = pngImg.In
						pngImg.In = pngImg.Out
						pngImg.Out = pngImg.Temp
					}
				} else if applyEffects[j] == "S" {
					if j == len(applyEffects)-1 {
						pngImg.Sharpen(true, -1, -1)
					} else {
						pngImg.Sharpen(false, -1, -1)
						pngImg.Temp = pngImg.In
						pngImg.In = pngImg.Out
						pngImg.Out = pngImg.Temp
					}
				} else if applyEffects[j] == "E" {
					if j == len(applyEffects)-1 {
						pngImg.EdgeDetection(true, -1, -1)
					} else {
						pngImg.EdgeDetection(false, -1, -1)
						pngImg.Temp = pngImg.In
						pngImg.In = pngImg.Out
						pngImg.Out = pngImg.Temp
					}
				} else if applyEffects[j] == "B" {
					if j == len(applyEffects)-1 {
						pngImg.Blur(true, -1, -1)
					} else {
						pngImg.Blur(false, -1, -1)
						pngImg.Temp = pngImg.In
						pngImg.In = pngImg.Out
						pngImg.Out = pngImg.Temp
					}
				}
			}
			err = pngImg.Save(fmt.Sprintf("../data/out/%v_%v", sizeList[a], effects[i]["inPath"]))
			if err != nil {
				panic(err)
			}
		}
	}
}
