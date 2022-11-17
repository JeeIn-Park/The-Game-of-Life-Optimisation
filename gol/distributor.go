package gol

import (
	"fmt"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func calculateNextAliveCells(turn int, p Params, world [][]byte, start int, finish int, c distributorChannels) []util.Cell {
	// find next alive cells from the given world
	var aliveCells []util.Cell

	for y := start; y < finish; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					//calculate the number of alive neighbour cells including itself
					if world[(y+i+p.ImageHeight)%p.ImageHeight][(x+j+p.ImageWidth)%p.ImageWidth] == 0xFF {
						sum++
					}
				}
			}

			var cell util.Cell
			cell.X, cell.Y = x, y

			// when the cell was alive in the given world, except it from the number of alive neighbour cells
			// then it keeps alive if it has 2 alive neighbours
			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, cell)
				} else if sum != 3 {
					c.events <- CellFlipped{
						CompletedTurns: turn,
						Cell:           cell,
					}
				}
			}

			// when a cell has three alive neighbours, it will be alive anyway
			if sum == 3 {
				aliveCells = append(aliveCells, cell)
				if world[y][x] == 0x00 {
					c.events <- CellFlipped{
						CompletedTurns: turn,
						Cell:           cell,
					}
				}
			}
		}
	}

	return aliveCells
}

func worldFromAliveCells(p Params, c []util.Cell) [][]byte {
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for _, i := range c {
		world[i.Y][i.X] = 0xFF
	}

	return world
}

func aliveCellWorker(turn int, p Params, world [][]byte, start int, finish int, cellOut chan<- []util.Cell, c distributorChannels) {
	cellPart := calculateNextAliveCells(turn, p, world, start, finish, c)
	cellOut <- cellPart
}

func writePgm(p Params, c distributorChannels, world [][]byte, turn int) {
	c.ioCommand <- ioOutput
	c.ioFilename <- fmt.Sprintf("%dx%dx%d", p.ImageHeight, p.ImageWidth, turn)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- world[y][x]
		}
	}
}

func quit(p Params, c distributorChannels, turn int, world [][]byte, aliveCells []util.Cell, ticker *time.Ticker) {
	// Report the final state using FinalTurnCompleteEvent.
	// - pass it down to events channel

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCells,
	}

	writePgm(p, c, world, turn)

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	ticker.Stop()
	close(c.events)

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	// Create a 2D slice to store the world.
	// - get the image in, so we can evolve it with the game of life algorithm (with IO goroutine)
	// - need to work out the file name from the parameter
	// - e.g. if we had two 256 by 256 coming in,
	//       we can make out a string and send that down via the appropriate channel
	//       after we've sent the appropriate command.
	//       we then get the image byte by byte and store it in this 2D world

	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
	}

	turn := 0
	var aliveCells []util.Cell

	// using ticker, report the number of cells that are still alive every 2 seconds
	// to report the count use the AliveCellsCount events.

	aliveCellsCount := 0

	// - need two 2D slices for this
	// - get final state of the world as it's evolved
	// Execute all turns of the Game of Life.
	// - for loop(call game of life function)

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCells = append(aliveCells, cell)
				aliveCellsCount = len(aliveCells)
			}
		}
	}
	world = worldFromAliveCells(p, aliveCells)

	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for range ticker.C {
			c.events <- AliveCellsCount{
				CompletedTurns: turn,
				CellsCount:     aliveCellsCount,
			}
		}
	}()
	for _, cell := range aliveCells {
		c.events <- CellFlipped{
			CompletedTurns: turn,
			Cell:           cell,
		}
	}
	pause := false

	go func() {
		for {
			keyPress := <-keyPresses
			switch keyPress {
			case 's':
				fmt.Println("writing pmg image")
				writePgm(p, c, world, turn)
			case 'q':
				fmt.Println("q is pressed, quit game of life")
				quit(p, c, turn, world, aliveCells, ticker)
			case 'p':
				func() {
					if pause == false {
						fmt.Println("Paused, current turn is", turn)
						pause = true
					} else if pause == true {
						fmt.Println("Continuing")
						pause = false
					}
				}()

			}
		}

	}()

	for i := 0; i < p.Turns; i++ {

		for pause {
		}
		if p.Threads == 1 {
			aliveCells = calculateNextAliveCells(turn, p, world, 0, p.ImageHeight, c)
			world = worldFromAliveCells(p, aliveCells)
			aliveCellsCount = len(aliveCells)
			c.events <- TurnComplete{CompletedTurns: turn}

		} else {
			aliveCells = []util.Cell{}
			size := (p.ImageHeight - (p.ImageHeight % p.Threads)) / p.Threads
			//remove possibility of remainder
			//remained world parts will be calculated at the last part

			cellOut := make([]chan []util.Cell, p.Threads)
			for k := range cellOut {
				cellOut[k] = make(chan []util.Cell)
			}

			for j := 0; j < p.Threads; j++ {
				if j == (p.Threads - 1) {
					go aliveCellWorker(turn, p, world, j*size, p.ImageHeight, cellOut[j], c)
				} else {
					go aliveCellWorker(turn, p, world, j*size, (j+1)*size, cellOut[j], c)
				}
			}

			for j := 0; j < p.Threads; j++ {
				cellPart := <-cellOut[j]
				aliveCells = append(aliveCells, cellPart...)
			}

			world = worldFromAliveCells(p, aliveCells)
			aliveCellsCount = len(aliveCells)
			c.events <- TurnComplete{CompletedTurns: turn}
		}

		turn++
	}

	quit(p, c, turn, world, aliveCells, ticker)
}
