package main

import (
	"github.com/esimov/colorquant"
	"image"
	"image/color/palette"
	"image/gif"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"
)

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
	log.Printf("dx1: %v, dx2: %v, dy1: %v, dy2: %v", dx1, dx2, dy1, dy2)
	for i := float64(1); i <= floatNumFrames; i++ {
		rects = append(rects, image.Rect(
			int(float64(origBounds.Min.X) + i * dx1),
			int(float64(origBounds.Min.Y) + i * dy1),
			int(float64(origBounds.Max.X) - i * dx2),
			int(float64(origBounds.Max.Y) - i * dy2),
			))
	}
	// TODO: append the reverse too for looping behavior
	// TODO: well actually, do that after the cropping / zooming to prevent duplication of work
	return rects
}

// TODO: share the logic between these 2 funcs
func reverseDelays(a []int) []int {
	out := make([]int, len(a))
	for i := len(a)-1; i >= 0; i-- {
		opp := len(a)-1-i
		out[i] = a[opp]
	}
	return out
}

func reverseImages(a []*image.Paletted) []*image.Paletted {
	out := make([]*image.Paletted, len(a))
	for i := len(a)-1; i >= 0; i-- {
		opp := len(a)-1-i
		out[i] = a[opp]
	}
	return out
}

// take an anim, and appends itself in reverse, creating a smooth loop
func loopAndReverse(anim gif.GIF) gif.GIF {
	var resolutions []string
	for _, res := range anim.Image {
		if res == nil {
			resolutions = append(resolutions, "nil")
		} else {
			resolutions = append(resolutions, res.Bounds().String())
		}
	}
	log.Printf("before reversing, resolutions: %s", strings.Join(resolutions, ", "))
	reversedImages := reverseImages(anim.Image)
	anim.Image = append(anim.Image, reversedImages...)
	reversedDelays := reverseDelays(anim.Delay)
	anim.Delay = append(anim.Delay, reversedDelays...)
	log.Printf("delays: %v", anim.Delay)
	return anim
}

func main() {
	inFile, err := os.Open(os.Args[1])
	if err != nil {
		panic("had trouble opening inFile: " + err.Error())
	}
	numFrames, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic("couldn't convert numFrames arg to int:  "+ err.Error())
	}
	defer inFile.Close()
	origImg, _, err := image.Decode(inFile)
	if err != nil {
		panic("had trouble decoding inFile: " + err.Error())
	}
	const delay = 5

	origQuantized := image.NewPaletted(origImg.Bounds(), palette.Plan9)
	floydSteinbergDitherer.Quantize(origImg, origQuantized, 256, true, true)
	anim := gif.GIF{LoopCount: numFrames * 2}
	anim.Image = append(anim.Image, origQuantized)
	anim.Delay = append(anim.Delay, delay)
	bestFaceRect, err := GetLargestFaceRect(origImg)
	if err != nil {
		panic("had trouble detecting faces in the image: " + err.Error())
	}
	scaledFaceBounds, err := getBoundsWithAspectRatio(origImg.Bounds(), bestFaceRect)
	if err != nil {
		panic("had trouble getting scaled bounds: " + err.Error())
	}
	rects := getIntermediateRects(origImg.Bounds(), scaledFaceBounds, numFrames / 2 - 1)
	log.Printf("we have %v rectangles", len(rects))

	for i, rect := range rects {
		log.Printf("rect #%v: %s", i, rect)
		croppedImg, err := Crop(origImg, rect)
		if err != nil {
			panic("had trouble cropping: " + err.Error())
		}
		resized := Resize(croppedImg, origImg.Bounds())

		quantized := image.NewPaletted(resized.Bounds(), palette.Plan9)
		floydSteinbergDitherer.Quantize(resized, quantized, 256, true, true)
		anim.Image = append(anim.Image, quantized)
		anim.Delay = append(anim.Delay, delay)
	}

	anim = loopAndReverse(anim)

	outFile, err := os.Create("/tmp/test.gif")
	if err != nil {
		panic("had trouble opening outFile: " + err.Error())
	}
	defer outFile.Close()

	err = gif.EncodeAll(outFile, &anim)
	if err != nil {
		panic("had trouble encoding outFile as gif: " + err.Error())
	}
}
