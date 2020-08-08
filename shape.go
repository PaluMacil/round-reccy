package main

import (
	"github.com/hajimehoshi/ebiten"
	"image"
	"image/color"
	"math"
)

type Point struct {
	X int
	Y int
}

func (p Point) Distance(p2 Point) float64 {
	return math.Sqrt(math.Pow(float64(p2.X-p.X), 2) + math.Pow(float64(p2.Y-p.Y), 2))
}

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
	// TODO: if border radius is > 50%, reduce to 50%; might need to handle ovals too

	// initialize boolean grid (no extra capacity in backing array is needed beyond initial length)
	b.cornerGrid = make([][]bool, b.BorderRadius+1, b.BorderRadius+1)
	for x := range b.cornerGrid {
		// add y column to slice
		b.cornerGrid[x] = make([]bool, b.BorderRadius+1, b.BorderRadius+1)
	}
	centerPt := Point{
		X: b.BorderRadius,
		Y: b.BorderRadius,
	}

	for x := 0; x <= b.BorderRadius; x++ {
		for y := 0; y <= b.BorderRadius; y++ {
			b.cornerGrid[x][y] = centerPt.Distance(Point{x, y}) <= float64(b.BorderRadius)
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
