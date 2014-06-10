package main

import (
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Fleet is a type to handle multiple ships
type Fleet struct {
	ships     map[*Ship]bool
	direction int
	stopCh    chan bool
	dirCh     chan int
	lock      sync.Mutex
}

func (fleet *Fleet) run() {
	direction := fleet.direction
	for {
		if len(fleet.ships) == 0 {
			cleanExit()
		}
		select {
		case direction = <-fleet.dirCh:
		case <-time.After(500000 * time.Microsecond):
			// Move all ships in the fleet and let them fire
			for ship, _ := range fleet.ships {
				ship.moveCh <- direction
				ship.fireFromFleet()
			}
		case <-fleet.stopCh:
			fleet.lock.Lock()
			for ship, _ := range fleet.ships {
				ship.stopCh <- true
			}
			return
		}
	}
}

// Initialize a fleet of enemy ships
func initFleet(size sizes, matrix [][]*Ship) {
	fleet := Fleet{
		ships:     make(map[*Ship]bool),
		direction: -1,
		stopCh:    make(chan bool),
		dirCh:     make(chan int),
	}

	midX := size.width / 2

	for j := 0; j < 3; j++ {
		for i := 0; i < 10; i++ {
			ship := &Ship{
				view:      [][]rune{{'(', '=', ')'}, {'\\', ' ', '/'}},
				positionX: midX - (-5+i)*5 - 5,
				positionY: 2 + 3*j,
				stopCh:    make(chan bool),
				moveCh:    make(chan int),
				fleet:     &fleet,
				size:      &size,
				matrix:    matrix,
				fg:        termbox.ColorYellow,
				bg:        termbox.ColorDefault,
			}
			go ship.run()
			fleet.ships[ship] = true
		}
	}
	go fleet.run()
}
