package main

import (
	"fmt"
	"github.com/esimov/colorquant"
	"image"
	"image/color/palette"
	"image/gif"
	_ "image/png"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const minLoggedDuration = time.Millisecond * 10 // we don't care about logging things that take <.01s

var floydSteinbergDitherer = colorquant.Dither{
	Filter: [][]float32{
		{0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0},
		{3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0},
		{1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0},
	},
}

func getIntermediateRects(origBounds, faceBounds image.Rectangle, nFrames int) []image.Rectangle {
	// it's nice to keep everything in floats and convert to int after all the math
	floatNumFrames := float64(nFrames)
	var rects []image.Rectangle
	dx1 := float64(faceBounds.Min.X - origBounds.Min.X) / floatNumFrames
	dx2 := float64(origBounds.Max.X - faceBounds.Max.X) / floatNumFrames
	dy1 := float64(faceBounds.Min.Y - origBounds.Min.Y) / floatNumFrames
	dy2 := float64(origBounds.Max.Y - faceBounds.Max.Y) / floatNumFrames
	for i := float64(1); i <= floatNumFrames; i++ {
		rects = append(rects, image.Rect(
			int(float64(origBounds.Min.X) + i * dx1),
			int(float64(origBounds.Min.Y) + i * dy1),
			int(float64(origBounds.Max.X) - i * dx2),
			int(float64(origBounds.Max.Y) - i * dy2),
			))
	}
	return rects
}

func panicIfError(err error, panicString string) {
	if err != nil {
		panic(panicString + ": " + err.Error())
	}
}

func logCheckpointTime(startTime time.Time, checkpoint *time.Duration, eventMsg string) {
	dur := time.Since(startTime) - *checkpoint
	if dur > minLoggedDuration {
		log.Printf("ran %s in %vs", eventMsg, dur.Seconds())
	}
	*checkpoint = time.Since(startTime)
}

// TODO: make this take a flag for finding the n best faces, and concatting the gifs
func main() {
	startTime := time.Now()
	log.Println("starting")
	if len(os.Args) != 3 {
		panic("usage: gif.go $IN_FILE $NUM_FRAMES")
	}
	inFile, err := os.Open(os.Args[1])
	panicIfError(err, "had trouble opening inFile")
	numFrames, err := strconv.Atoi(os.Args[2])
	panicIfError(err, "couldn't convert numFrames arg to int")
	defer inFile.Close()
	origImg, _, err := image.Decode(inFile)
	panicIfError(err, "had trouble decoding inFile")
	const delay = 5

	checkpoint := time.Since(startTime)

	origQuantized := image.NewPaletted(origImg.Bounds(), palette.Plan9)
	floydSteinbergDitherer.Quantize(origImg, origQuantized, 256, true, true)
	logCheckpointTime(startTime, &checkpoint, "quantization / dithering of input image")
	//colorquant.NoDither.Quantize(origImg, origQuantized, 256, false, true)
	if reflect.DeepEqual(origQuantized.Palette, palette.Plan9) {
		log.Printf("the quantized image still has the Plan9 palette! SAD!")
	}
	anim := gif.GIF{LoopCount: numFrames} // TODO: multiply this by numFaces
	anim.Image = make([]*image.Paletted, numFrames)
	anim.Delay = make([]int, numFrames)
	// put the original image at the start and end
	anim.Image[0] = origQuantized
	anim.Image[numFrames - 1] = origQuantized
	for i := 0; i < numFrames; i++ {
		anim.Delay[i] = delay
	}

	bestFaceRect, err := GetBestFaceRect(origImg)
	logCheckpointTime(startTime, &checkpoint, "face detection")
	panicIfError(err, "had trouble detecting faces in the image")
	scaledFaceBounds, err := getBoundsWithAspectRatio(origImg.Bounds(), bestFaceRect)
	panicIfError(err, "had trouble getting scaled bounds")
	rects := getIntermediateRects(origImg.Bounds(), scaledFaceBounds, numFrames / 2 - 1)

	checkpoint = time.Since(startTime)
	wg := new(sync.WaitGroup)
	cropResults := make(chan CropResult, len(rects))
	for i, rect := range rects {
		wg.Add(1)
		go cropAndResize(&cropResults, wg, i, numFrames, rect, origQuantized)
	}
	go func(wg *sync.WaitGroup, results chan CropResult) {
		wg.Wait()
		close(results)
	}(wg, cropResults)
	for result := range cropResults {
		for _, index := range result.indices {
			anim.Image[index] = result.img
		}
	}
	logCheckpointTime(startTime, &checkpoint, "concurrently created intermediate images")

	outFile, err := os.Create("/tmp/test.gif")
	panicIfError(err, "had trouble opening outFile")
	defer outFile.Close()

	err = gif.EncodeAll(outFile, &anim)
	logCheckpointTime(startTime, &checkpoint, "created and encoded gif file")
	panicIfError(err, "had trouble encoding outFile as gif")
	log.Printf("finished in %vs", time.Since(startTime).Seconds())
}

type CropResult struct {
	// the indices in the gif in which to place the cropped / resized image
	indices []int
	img *image.Paletted
}

func cropAndResize(
	results *chan CropResult,
	wg *sync.WaitGroup,
	origIdx,
	totalFrames int,
	cropTo image.Rectangle,
	origImg *image.Paletted) {
	defer wg.Done()
	funcStart := time.Now()
	log.Printf("rect #%v: %s", origIdx, cropTo)
	croppedImg, err := Crop(origImg, cropTo)
	checkpoint := time.Since(funcStart)
	logCheckpointTime(funcStart, &checkpoint, fmt.Sprintf("crop #%v", origIdx))
	panicIfError(err, "had trouble cropping")
	resized := Resize(croppedImg, origImg.Bounds())
	logCheckpointTime(funcStart, &checkpoint, fmt.Sprintf("resize #%v", origIdx))
	// we already have the full size image at the first and last index
	reverseIndex := totalFrames - 2 - origIdx
	*results <- CropResult{indices: []int{origIdx + 1, reverseIndex}, img: resized}
	log.Printf("ran cropAndResize for img #%v in %vs", origIdx, time.Since(funcStart).Seconds())
}
