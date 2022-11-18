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

func makeCall(client *rpc.Client, world [][]byte, turn int, imageHeight int, imageWidth int) [][]byte {

	fmt.Println("makeCall is called")
	request := stubs.Request{
		InitialWorld: world,
		Turn:         turn,
		ImageHeight:  imageHeight,
		ImageWidth:   imageWidth,
	}
	fmt.Println("request is created")
	response := new(stubs.Response)
	fmt.Println("response is created")
	client.Call(stubs.EvaluateAllHandler, request, response)
	fmt.Println("client call done")

	return response.FinalWorld
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	//flag.Parse()

	server := "127.0.0.1:8030"
	//client, _ := rpc.Dial("tcp", *server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()
	fmt.Println("client: ready to send")

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
	fmt.Println("initial world is ready")

	finalWorld := makeCall(client, world, p.Turns, p.ImageHeight, p.ImageWidth)
	fmt.Println("got finalWorld from the server")
	var aliveCell []util.Cell
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if finalWorld[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCell = append(aliveCell, cell)
			}
		}
	}

	fmt.Println("ready to terminate")
	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          aliveCell,
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
