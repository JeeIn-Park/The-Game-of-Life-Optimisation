package gol

import (
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

func calculateNextAliveCells(p Params, world [][]byte) []util.Cell {
	// takes the current state of the world and completes one evolution of the world
	// find next alive cells calculating each cell in the given world
	var aliveCells []util.Cell

	for y := 0; y < p.ImageHeight; y++ {
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

			// when the cell was alive in the given world, exclude it from the number of alive neighbour cells
			// then it keeps alive if it has 2 alive neighbours
			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, cell)
				}
			}

			// when a cell has three alive neighbours, it will be alive anyway
			if sum == 3 {
				aliveCells = append(aliveCells, cell)
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// Create a 2D slice to store the world.
	// - get the image in, so we can evolve it with the game of life algorithm (with IO goroutine)
	// - need to work out the file name from the parameter
	// - eg. if we had two 256 by 256 coming in,
	//       we can make out a string and send that down via the appropriate channel
	//       after we've sent the appropriate command.
	//       we then get the image byte by byte and store it in this 2D world
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	var aliveCells []util.Cell
	turn := 0

	// Execute all turns of the Game of Life.
	// - for loop(call game of life function)
	// - get final state of the world as it's evolved
	// - need two 2D slices for this
	for i := 0; i < p.Turns; i++ {
		aliveCells = calculateNextAliveCells(p, world)
		world = worldFromAliveCells(p, aliveCells)
	}

	// Report the final state using FinalTurnCompleteEvent.
	// - pass it down to events channel

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          aliveCells,
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
