package gol

import (
	"fmt"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
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

func aliveCellFromWorld(p Params, world [][]byte) []util.Cell {
	var aliveCell []util.Cell
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCell = append(aliveCell, cell)
			}
		}
	}
	return aliveCell
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

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	//flag.Parse()

	server := "127.0.0.1:8030"
	//client, _ := rpc.Dial("tcp", *server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()

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

	request := stubs.Request{
		InitialWorld: world,
		Turn:         p.Turns,
		ImageHeight:  p.ImageHeight,
		ImageWidth:   p.ImageWidth,
	}
	response := new(stubs.Response)

	client.Call(stubs.EvaluateAllHandler, request, response)

	aliveCell := aliveCellFromWorld(p, response.ComputedWorld)

	c.events <- FinalTurnComplete{
		CompletedTurns: response.CompletedTurn,
		Alive:          aliveCell,
	}

	writePgm(p, c, response.ComputedWorld, response.CompletedTurn)

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
