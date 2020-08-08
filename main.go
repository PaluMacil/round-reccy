package main

import (
	"github.com/hajimehoshi/ebiten"
	"log"
)

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Rectangles!!!")
	log.Println("Starting Game!")
	g := &Game{}
	g.Watch("shapes.json")
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(g); err != nil {
		g.Stop()
		log.Fatal(err)
	}
	g.Stop()
}
