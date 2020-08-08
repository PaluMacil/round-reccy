package main

import (
	"github.com/hajimehoshi/ebiten"
	"gonum.org/v1/plot/tools/bezier"
	"gonum.org/v1/plot/vg"
	"image"
	"image/color"
	"math"
)

type Blocks []Block

type Block struct {
	Left         int        `json:"left"`
	Top          int        `json:"top"`
	Height       int        `json:"height"`
	Width        int        `json:"width"`
	BorderWidth  int        `json:"borderWidth"`
	BorderColor  color.RGBA `json:"borderColor"`
	BorderRadius int        `json:"borderRadius"`
	FillColor    color.RGBA `json:"fillColor"`
	cornerGrid   [][]bool
}

func (b *Block) calcCornerGrid() {
	// TODO: if border radius is > 50%, reduce to 50%; might need to handle individual corners first

	// initialize boolean grid (no extra capacity in backing array is needed beyond initial length)
	// TODO: Is this is much too large?
	b.cornerGrid = make([][]bool, b.Width, b.Width)
	for i := range b.cornerGrid {
		b.cornerGrid[i] = make([]bool, b.Height, b.Height)
	}
	curve := bezier.New(
		vg.Point{Y: vg.Length(b.BorderRadius)},
		vg.Point{X: vg.Length(b.BorderRadius)},
		vg.Point{},
	)
	for x := 0; x <= b.BorderRadius; x++ {
		// Get point on bezier
		pt := curve.Point(float64(x) / float64(b.BorderRadius))
		// start at tallest point within curve for this x and decrease to end of grid
		highestY := int(math.Round(float64(pt.Y)))

		for y := highestY; y <= b.BorderRadius; y++ {
			b.cornerGrid[x][y] = true
		}
	}
}

func (b Block) IsInside(x, y int) bool {
	if b.BorderRadius <= 0 {
		return true
	}
	// test case: take a notch out of the rectangle
	/*if x < 25 && y < 25 {
		return false
	}*/
	// left column
	if x <= b.BorderRadius {
		// top left
		if y <= b.BorderRadius {
			return b.cornerGrid[x][y]
		}
		// bottom left
		if y >= b.Height-b.BorderRadius {
			return b.cornerGrid[x][b.Height-y]
		}
		// middle row remains
	}
	// right column
	if x >= b.Width-b.BorderRadius {
		// top right
		if y <= b.BorderRadius {
			return b.cornerGrid[b.Width-x][y]
		}
		// bottom right
		if y >= b.Height-b.BorderRadius {
			return b.cornerGrid[b.Width-x][b.Height-y]
		}
		// middle row remains
	}

	// middle column remains (x > b.BorderRadius && x < b.Width - b.BorderRadius) along with middle row (above)
	return true
}

func (b Block) Mask() *ebiten.Image {
	if b.BorderRadius > 0 {
		b.calcCornerGrid()
	}
	alphas := image.NewAlpha(image.Rectangle{
		Min: image.Point{},
		Max: image.Point{X: b.Width, Y: b.Height},
	})
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			if b.IsInside(x, y) {
				alphas.SetAlpha(x, y, color.Alpha{A: 0xFF})
			} else {
				alphas.SetAlpha(x, y, color.Alpha{A: 0x00})
			}
		}
	}
	mask, _ := ebiten.NewImageFromImage(alphas, ebiten.FilterDefault)

	return mask
}
