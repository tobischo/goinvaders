package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/nsf/termbox-go"
)

type sizes struct {
	width  int
	height int
}

func initMatrix(w, h int) [][]*Ship {
	matrix := make([][]*Ship, w+1)
	for x := 0; x < w+1; x++ {
		matrix = append(matrix, make([]*Ship, h+1))
		for y := 0; y < h+1; y++ {
			matrix[x] = append(matrix[x], nil)
		}
	}
	return matrix
}

func cleanExit() {
	termbox.Close()
	os.Exit(0)
}

func main() {
	err := termbox.Init()
	if err != nil {
		os.Exit(1)
	}
	termbox.HideCursor()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	w, h := termbox.Size()
	size := sizes{width: w, height: h}

	matrix := initMatrix(w, h)

	fpsSleepTime := time.Duration(1000000/30) * time.Microsecond
	go func() {
		for {
			time.Sleep(fpsSleepTime)
			termbox.Flush()
		}
	}()

	eventChan := make(chan termbox.Event)
	go func() {
		for {
			event := termbox.PollEvent()
			eventChan <- event
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	ship := initShip(size, matrix)
	go ship.run()

	initBarrier(size, matrix)
	initFleet(size, matrix)

eventLoop:
	for {
		select {
		case event := <-eventChan:
			switch event.Type {
			case termbox.EventKey:
				switch event.Key {
				case termbox.KeyCtrlZ, termbox.KeyCtrlC:
					break eventLoop
				case termbox.KeyArrowLeft:
					ship.moveLeft()
				case termbox.KeyArrowRight:
					ship.moveRight()
				case termbox.KeySpace:
					ship.fire()
				}

				switch event.Ch {
				case 'q':
					break eventLoop
				case 'c':
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				}
			case termbox.EventError:
				break eventLoop
			}
		case <-sigChan:
			break eventLoop
		}
	}

	cleanExit()
}
