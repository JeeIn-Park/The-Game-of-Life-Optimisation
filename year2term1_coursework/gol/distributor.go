package gol

import (
	"fmt"
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

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	world = worldFromAliveCells(p, aliveCells)

	aliveCells, turn = <-response
	world = worldFromAliveCells(p, aliveCells)

	// fmt.Println("Responded:" + response.AliveCells) -> 이거 대신

	// Report the final state using FinalTurnCompleteEvent.
	// - pass it down to events channel

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCells, // 확인
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)

}
