package main

import (
	pigo "github.com/esimov/pigo/core"
	"image"
	"io/ioutil"
	"log"
	"sort"
)

// TODO: make this return a slice of the N best face rects
func GetBestFaceRect(img image.Image) (image.Rectangle, error) {
	//cascade, err := ioutil.ReadFile("/var/www/prettygood.dev/cascade/facefinder")
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

	scoresToFaces := getScoresToFaceRectangles(faces)
	var scores []float64
	for k := range scoresToFaces {
		scores = append(scores, k)
	}
	sort.Float64s(scores)
	return scoresToFaces[scores[len(scores) - 1]], nil
}

func getScoresToFaceRectangles(faceDetections []pigo.Detection) map[float64]image.Rectangle {
	outMap := make(map[float64]image.Rectangle)
	for _, face := range faceDetections {
		rect := image.Rect(
			face.Col-face.Scale/2,
			face.Row-face.Scale/2,
			face.Col+face.Scale/2,
			face.Row+face.Scale/2,
		)
		log.Printf("found a face with dims: %s, score: %v", rect.String(), face.Q)
		// let's try making score the detection score * area
		score := float64(face.Q) * float64(rect.Dx() * rect.Dy())
		outMap[score] = rect
	}
	return outMap

}
