// handles the cropping and resizing of images

package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"log"
)

var blankImage = image.NewRGBA(image.Rectangle{})

func getBoundsWithAspectRatio(oldBounds, newBounds image.Rectangle) (image.Rectangle, error) {
	// assert that newBounds is inside oldBounds
	if !newBounds.In(oldBounds) {
		return oldBounds, fmt.Errorf("newBounds not within bounds of original image")
	}
	oldAspectRatio := float64(oldBounds.Dy()) / float64(oldBounds.Dx())
	newAspectRatio := float64(newBounds.Dy()) / float64(newBounds.Dx())

	log.Printf("oldAspectRatio: %f, newAspectRatio: %f", oldAspectRatio, newAspectRatio)
	if oldAspectRatio == newAspectRatio {
		return newBounds, nil
	}

	var scaledNewBounds image.Rectangle
	// get the bigger side to keep
	// naively scale it for now, worry about edge detection later
	if newAspectRatio > oldAspectRatio {
		// use y from newBounds and scale x to preserve aspect ratio
		xShift := int((newAspectRatio / oldAspectRatio * float64(newBounds.Dx()) - float64(newBounds.Dx())) / 2)
		log.Printf("xShift: %v", xShift)
		scaledNewBounds = image.Rect(
			newBounds.Min.X - xShift,
			newBounds.Min.Y,
			newBounds.Max.X + xShift,
			newBounds.Max.Y)
	} else {
		// use x from newBounds and scale y to preserve aspect ratio
		yShift := int((oldAspectRatio / newAspectRatio * float64(newBounds.Dy()) - float64(newBounds.Dy())) / 2)
		scaledNewBounds = image.Rect(
			newBounds.Min.X,
			newBounds.Min.Y - yShift,
			newBounds.Max.X,
			newBounds.Max.Y + yShift)
	}
	log.Printf("scaledNewBounds before shift: %s", scaledNewBounds)
	// now we shift the scaledNewBounds if they aren't fully enclosed in the original rect
	if scaledNewBounds.In(oldBounds) {
		return scaledNewBounds, nil
	} else {
		// the scaledNewBounds overlap with exactly one edge
		// find that edge and shift by the difference, using Add
		var shiftBy image.Point
		if scaledNewBounds.Min.X < oldBounds.Min.X {
			shiftBy = image.Point{X: oldBounds.Min.X - scaledNewBounds.Min.X, Y: 0}
		} else if scaledNewBounds.Max.X > oldBounds.Max.X {
			shiftBy = image.Point{X: oldBounds.Max.X - scaledNewBounds.Max.X, Y: 0}
		} else if scaledNewBounds.Min.Y < oldBounds.Min.Y {
			shiftBy = image.Point{X: 0, Y: oldBounds.Min.Y - scaledNewBounds.Min.Y}
		} else if scaledNewBounds.Max.Y > oldBounds.Max.Y {
			shiftBy = image.Point{X: 0, Y: oldBounds.Max.Y - scaledNewBounds.Max.Y}
		}

		return scaledNewBounds.Add(shiftBy), nil
	}
}

func Crop(img image.Image, newBounds image.Rectangle) (image.Image, error) {
	bounds := img.Bounds()
	log.Printf("the bounds of the original image are %s", bounds)

	if !newBounds.In(bounds) {
		return blankImage, fmt.Errorf("newBounds not within bounds of original image, newBounds: " + newBounds.String())
	}

	//newImage := image.NewPaletted(newBounds, palette.Plan9)
	newImage := image.NewRGBA(newBounds)
	draw.Draw(newImage, newImage.Bounds(), img, newBounds.Min, draw.Src)

	log.Printf("the bounds of the new image are %s", newBounds)

	return newImage, nil
}

func Resize(img image.Image, bounds image.Rectangle) image.Image{
	return resize.Resize(uint(bounds.Dx()), uint(bounds.Dy()), img, resize.Lanczos2)
}

//func main() {
//	inFile, err := os.Open(os.Args[1])
//	if err != nil {
//		panic("had trouble opening inFile")
//	}
//	defer inFile.Close()
//	img, _, err := image.Decode(inFile)
//
//	unscaledCropBounds := image.Rect(10, 10, 100, 200)
//	cropBounds, err := getBoundsWithAspectRatio(img.Bounds(), unscaledCropBounds)
//	if err != nil {
//		panic("had trouble getting scaled bounds")
//	}
//	croppedImg, err := Crop(img, cropBounds)
//	if err != nil {
//		panic("had trouble cropping")
//	}
//
//	//resizedImg := Resize(croppedImg, img.Bounds())
//	resizedImg := croppedImg
//
//	outFile, err := os.Create("/tmp/cropped2.png")
//	if err != nil {
//		panic("had trouble opening outFile")
//	}
//	defer outFile.Close()
//
//	err = png.Encode(outFile, resizedImg)
//	if err != nil {
//		panic("had trouble encoding outFile as png")
//	}
//
//}

