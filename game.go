package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"image/color"
	"log"
	"os"
	"sync"
)

type Game struct {
	sync.RWMutex
	watcher *fsnotify.Watcher
	shapes  Blocks
}

func (g *Game) SetShapes(shapes Blocks) {
	g.Lock()
	defer g.Unlock()
	g.shapes = shapes
}

func (g *Game) Shapes() []Block {
	g.RLock()
	defer g.RUnlock()
	return g.shapes
}

func (g *Game) Update(screen *ebiten.Image) error {

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Get the x, y position of the cursor from the CursorPosition() function
	x, y := ebiten.CursorPosition()

	// Display the information with "X: xx, Y: xx" format
	err := ebitenutil.DebugPrint(screen, fmt.Sprintf("X: %d, Y: %d", x, y))
	if err != nil {
		panic(err)
	}

	shapes := g.Shapes()
	for _, shape := range shapes {
		finalImg, _ := ebiten.NewImage(shape.Width, shape.Height, ebiten.FilterNearest)
		// this can be an optimization if I'm keeping it around
		// I will only want to redraw final image if something about it changes
		// finalImg.Fill(color.White)

		// original filled rectangle before border, radius, etc are applied
		rectImg, _ := ebiten.NewImage(shape.Width, shape.Height, ebiten.FilterNearest)
		_ = rectImg.Fill(shape.FillColor)
		// mask image
		maskImg := shape.Mask()

		// copy rectImg and maskImg with source-in composite mode to use the source's
		// alpha channel's opacity values (draw mask to image
		maskOpt := &ebiten.DrawImageOptions{}
		maskOpt.CompositeMode = ebiten.CompositeModeCopy
		finalImg.Fill(color.White)
		finalImg.DrawImage(maskImg, maskOpt)
		copyOpt := &ebiten.DrawImageOptions{}
		copyOpt.CompositeMode = ebiten.CompositeModeSourceIn
		finalImg.DrawImage(rectImg, copyOpt)

		// copy final image to screen with geo transform
		finalOpt := &ebiten.DrawImageOptions{}
		// TODO: right now copy overwrites destination alpha, but copy is fine for now
		finalOpt.CompositeMode = ebiten.CompositeModeSourceOver //Copy
		finalOpt.GeoM.Translate(float64(shape.Left), float64(shape.Top))
		screen.DrawImage(finalImg, finalOpt)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (g *Game) Watch(filename string) {
	var err error
	// initial load of shapes before watches
	g.loadShapes(filename)
	g.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("cannot watch filename: %s", err)
	}
	go func() {
		for {
			select {
			case event, ok := <-g.watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					g.loadShapes(filename)
				}
			case err, ok := <-g.watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = g.watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("file watch start-up complete")
}

func (g *Game) Stop() error {
	return g.watcher.Close()
}

func (g *Game) loadShapes(filename string) {
	fmt.Println("opening", filename)
	shapesFile, err := os.Open(filename)
	if err != nil {
		log.Printf("could not open shapes filename: %s\n", err)
	}
	jsonParser := json.NewDecoder(shapesFile)
	var newShapes Blocks
	if err = jsonParser.Decode(&newShapes); err != nil {
		log.Printf("parsing shapes filename: %s\n", err)
	}
	fmt.Println("found", len(newShapes), "shapes")
	g.SetShapes(newShapes)
}
