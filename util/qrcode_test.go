package util

import (
	"image/color"
	"testing"
)

func TestColor(t *testing.T) {

	v := color.Color(color.RGBA{R: 255, G: 255, B: 255, A: 255}) == color.Color(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	t.Log(v)

}
