package main

import (
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	left  int = -1
	down  int = 0
	right int = 1
)

// Ship that protects the planet
type Ship struct {
	view      [][]rune
	positionX int
	positionY int
	stopCh    chan bool
	moveCh    chan int
	fleet     *Fleet
	size      *sizes
	matrix    [][]*Ship
	control   bool
	fg        termbox.Attribute
	bg        termbox.Attribute
}

// Create the player controller ship
func initShip(size sizes, matrix [][]*Ship) Ship {
	return Ship{
		view:      [][]rune{{'|'}, {'<', '=', '>'}},
		positionX: size.width / 2,
		positionY: size.height,
		stopCh:    make(chan bool),
		moveCh:    make(chan int),
		size:      &size,
		matrix:    matrix,
		control:   true,
		fg:        termbox.ColorWhite,
		bg:        termbox.ColorDefault,
	}
}

// Create a group of not moving ships which act as a barrier
func initBarrier(size sizes, matrix [][]*Ship) {
	positions := []int{0, 1, 1, 1, 0}

	for i := 3; i < size.width-4; i = i + 7 {
		for count, value := range positions {
			for j := 0; j < 2; j++ {
				ship := Ship{
					view:      [][]rune{{' '}},
					positionX: i + count,
					positionY: size.height - 4 - value + j,
					stopCh:    make(chan bool),
					moveCh:    make(chan int),
					size:      &size,
					matrix:    matrix,
					fg:        termbox.ColorDefault,
					bg:        termbox.ColorGreen,
				}
				go ship.run()
			}
		}
	}
}

func (ship *Ship) run() {
	ship.draw()
	for {
		select {
		case <-ship.stopCh:
			// Stop the ship and remove it from the fleet if it is affiliated
			if ship.fleet != nil {
				delete(ship.fleet.ships, ship)
			}
			ship.clear()
			return
		case dir := <-ship.moveCh:
			switch dir {
			case left:
				if ship.positionX > 1 {
					ship.move(left)
				}

				if ship.fleet != nil && ship.positionX == 1 {
					ship.fleet.dirCh <- down
				}
			case down:
				if ship.positionY < ship.size.height {
					ship.move(down)
				}

				if ship.fleet != nil {
					if ship.positionX == 1 {
						ship.fleet.dirCh <- right
					}
					if ship.positionX == ship.size.width-2 {
						ship.fleet.dirCh <- left
					}

					// Stop if fleet reaches defense ship
					if ship.positionY == ship.size.height-1 {
						ship.fleet.stopCh <- true
						cleanExit()
					}
				}
			case right:
				if ship.positionX < ship.size.width-2 {
					ship.move(right)
				}

				if ship.fleet != nil && ship.positionX == ship.size.width-2 {
					ship.fleet.dirCh <- down
				}
			}
		case <-time.After(1000000 / 30 * time.Microsecond):
			ship.draw()
		}
	}
}

// Fire a shot from a ship - directed upwards or downwards depending on
// fleet affiliation
func (ship *Ship) fire() {
	var dir, offset int
	var color termbox.Attribute
	// Avoid creating the shot _in_ the ship
	if ship.fleet != nil {
		dir = 1
		offset = 0
		color = termbox.ColorRed
	} else {
		dir = -1
		offset = -3
		color = termbox.ColorCyan
	}

	shot := Shot{
		positionX: ship.positionX,
		positionY: ship.positionY + offset,
		direction: dir,
		stopCh:    make(chan bool),
		ship:      ship,
		fg:        color,
		bg:        termbox.ColorDefault,
	}

	go shot.run()
}

func (ship *Ship) fireFromFleet() {
	r := rand.Intn(100)
	if r < 5 && ship.positionY+6 < ship.size.height {
		matrix := ship.matrix
		ship_3 := matrix[ship.positionX][ship.positionY+2]
		ship_6 := matrix[ship.positionX][ship.positionY+5]
		if !((ship_3 != nil && ship_3.fleet != nil) ||
			(ship_6 != nil && ship_6.fleet != nil)) {
			ship.fire()
		}
	}
}

// Draw the ship and set the reference in the collision matrix
func (ship *Ship) draw() {
	rows := len(ship.view)
	matrix := ship.matrix

	for r, row := range ship.view {
		width := len(row) / 2
		for c, ch := range row {
			x := ship.positionX - width + c
			y := ship.positionY - rows + r
			if matrix[x][y] != nil && matrix[x][y] != ship {
				matrix[x][y].stopCh <- true
			}
			matrix[x][y] = ship
			termbox.SetCell(x, y, ch, ship.fg, ship.bg)
		}
	}
}

// Make sure all view fields of the ship are cleared as well as
// the collision matrix
func (ship *Ship) clear() {
	rows := len(ship.view)
	matrix := ship.matrix

	for r, row := range ship.view {
		width := len(row) / 2
		for c := range row {
			x := ship.positionX - width + c
			y := ship.positionY - rows + r
			matrix[x][y] = nil
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func (ship *Ship) moveLeft() {
	ship.moveCh <- -1
}

func (ship *Ship) moveRight() {
	ship.moveCh <- 1
}

func (ship *Ship) moveDown() {
	ship.moveCh <- 0
}

func (ship *Ship) move(direction int) {
	ship.clear()
	if direction == down {
		ship.positionY++
	} else {
		ship.positionX += direction
	}
	ship.draw()
}
