package main

import (
	"time"

	"github.com/nsf/termbox-go"
)

//Shot fired by a ship
type Shot struct {
	positionX int
	positionY int
	direction int
	stopCh    chan bool
	ship      *Ship
	fg        termbox.Attribute
	bg        termbox.Attribute
}

func (shot *Shot) run() {
	shot.draw()
	for {
		select {
		case <-time.After(250000 * time.Microsecond):
			shot.clear()
			shot.positionY += shot.direction
			shot.draw()
			shot.detectCollision()
		case <-shot.stopCh:
			shot.clear()
			return
		}
	}
}

func (shot *Shot) draw() {
	if shot.positionY < 0 || shot.positionY > shot.ship.size.height {
		shot.stopCh <- true
	}

	termbox.SetCell(
		shot.positionX,
		shot.positionY,
		'|',
		shot.fg,
		shot.bg,
	)
}

func (shot *Shot) detectCollision() {
	// Avoid checking if the shot has reached the top corner
	if shot.positionY >= 0 {
		ship := shot.ship.matrix[shot.positionX][shot.positionY]
		// Check if there is a ship (which did not fire the shot)
		if ship != nil && ship != shot.ship {
			// Check for player controlled ship
			if ship.control {
				cleanExit()
			} else {
				ship.stopCh <- true
				shot.clear()
				shot.stopCh <- true
			}
		}
	}
}

func (shot *Shot) clear() {
	termbox.SetCell(
		shot.positionX,
		shot.positionY,
		' ',
		termbox.ColorDefault,
		termbox.ColorDefault,
	)
}
