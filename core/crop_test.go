package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"image"
	"testing"
)

func TestGetBoundsWithAspectRatio(t *testing.T) {
	t.Run("simplest case, newBounds happen to be in scale", func(t *testing.T) {
		orig := image.Rect(0, 0, 200, 100)
		newBounds := image.Rect(40, 40, 60, 50)
		got, err := getBoundsWithAspectRatio(orig, newBounds)
		assert.Equal(t, newBounds, got) // no change since they're same aspect ratio
		assert.Nil(t, err)
	})

	t.Run("simple case, no edge collision", func(t *testing.T) {
		orig := image.Rect(0, 0, 200, 100)
		newBounds := image.Rect(40, 40, 50, 50)
		want := image.Rect(35, 40, 55, 50)
		got, err := getBoundsWithAspectRatio(orig, newBounds)
		assert.Equal(t, want, got)
		assert.Nil(t, err)
	})

	t.Run("scale along an x edge, shift as expected", func(t *testing.T) {
		orig := image.Rect(0, 0, 200, 100)
		newBounds := image.Rect(0, 40, 10, 50)
		want := image.Rect(0, 40, 20, 50)
		got, err := getBoundsWithAspectRatio(orig, newBounds)
		assert.Equal(t, want, got)
		assert.Nil(t, err)
	})

	t.Run("scale along a far y edge, shift as expected", func(t *testing.T) {
		orig := image.Rect(100, 100, 200, 500)
		newBounds := image.Rect(150, 485, 160, 495)
		want := image.Rect(150, 460, 160, 500)
		got, err := getBoundsWithAspectRatio(orig, newBounds)
		assert.Equal(t, want, got)
		assert.Nil(t, err)
	})

	t.Run("newBounds outside orig", func(t *testing.T) {
		orig := image.Rect(100, 100, 300, 200)
		newBounds := image.Rect(0, 40, 10, 50)
		_, err := getBoundsWithAspectRatio(orig, newBounds)
		assert.Equal(t, err, fmt.Errorf("newBounds not within bounds of original image"))
	})
}
