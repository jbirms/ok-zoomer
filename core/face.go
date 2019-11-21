package main

import (
	pigo "github.com/esimov/pigo/core"
	//"github.com/esimov/pigo/wasm/detector"
	"image"
	"io/ioutil"
	"log"
)

func GetLargestFaceRect(img image.Image) (image.Rectangle, error) {
	//det := detector.NewDetector()

	//cascade, err := det.FetchCascade("https://raw.githubusercontent.com/esimov/pigo/master/cascade/facefinder")
	cascade, err := ioutil.ReadFile("../cascade/facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
		return image.Rectangle{}, err
	}

	ngrbaImg := pigo.ImgToNRGBA(img)
	pixels := pigo.RgbToGrayscale(ngrbaImg)
	cols, rows := ngrbaImg.Bounds().Max.X, ngrbaImg.Bounds().Max.Y

	// these are the defaults from the pigo README
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	pg := pigo.NewPigo()
	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	classifier, err := pg.Unpack(cascade)
	if err != nil {
		log.Fatalf("Error reading the cascade file: %s", err)
		return image.Rectangle{}, err
	}

	angle := 0.0 // cascade rotation angle. 0.0 is 0 radians and 1.0 is 2*pi radians

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	dets := classifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	faces := classifier.ClusterDetections(dets, 0.2)
	log.Printf("detected %v faces!", len(faces))

	return getBestFace(faces), nil
}

// for now let's just pick the biggest rectangle, but maybe we'll incorporate the detection score (Detection.Q)
func getBestFace(faceDetections []pigo.Detection) image.Rectangle {
	var bestFaceRect image.Rectangle
	var hiScore float32
	for _, face := range faceDetections {
		//rect := image.Rect(
		//	face.Col-face.Scale/2,
		//	face.Row-face.Scale/2,
		//	face.Scale,
		//	face.Scale,
		//)
		rect := image.Rect(
			face.Col-face.Scale/2,
			face.Row-face.Scale/2,
			face.Col+face.Scale/2,
			face.Row+face.Scale/2,
		)
		log.Printf("found a face with dims: %s, score: %v", rect.String(), face.Q)
		//if rect.Dx() * rect.Dy() > bestFaceRect.Dx() * bestFaceRect.Dy() {
		if face.Q > hiScore {
			bestFaceRect = rect
		}
	}
	log.Printf("the best face in the image has dimensions: %s", bestFaceRect.String())
	return bestFaceRect

}
