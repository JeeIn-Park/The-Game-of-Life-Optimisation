package gol

import (
	"fmt"
	"net/rpc"
	"time"
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

func quit(p Params, c distributorChannels, turn int, world [][]byte, aliveCells []util.Cell, ticker *time.Ticker) {
	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCells,
	}
	writePgm(p, c, world, turn)
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	ticker.Stop()
	close(c.events)

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keypress <-chan rune) {
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
	var aliveCell []util.Cell
	var turn int

	done := make(chan *rpc.Call, 10)
	client.Go(stubs.EvaluateAllHandler, request, response, done)

	//client.Call(stubs.EvaluateAllHandler, request, response)
	ticker := time.NewTicker(time.Second * 2)
	go func() {
		for range ticker.C {
		}
	}()

	<-done
	aliveCell = aliveCellFromWorld(p, response.ComputedWorld)
	turn = response.CompletedTurn

	quit(p, c, turn, world, aliveCell, ticker)
}
