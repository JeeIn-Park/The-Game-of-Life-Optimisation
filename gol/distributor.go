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

func makeCall(client *rpc.Client, world [][]byte, turn int, imageHeight int, imageWidth int, c distributorChannels) [][]byte {

	request := stubs.Request{
		InitialWorld: world,
		Turn:         turn,
		ImageHeight:  imageHeight,
		ImageWidth:   imageWidth,
	}
	response := new(stubs.Response)
	client.Call(stubs.EvaluateAllHandler, request, response)

	return response.FinalWorld
}

func aliveCellsFromWorld(world [][]byte, imageHeight int, imageWidth int) []util.Cell {
	var aliveCell []util.Cell
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCell = append(aliveCell, cell)
			}
		}
	}
	return aliveCell
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

	// might have a thread on the local controller that pause
	//the aws node and get the alive cell count every two seconds

	// might have some two way RPC setup
	//where the aws node calls back to the local controller
	//and then it reports the life cell count

	ticker := time.NewTicker(time.Second * 2)
	fmt.Println("ticker is created")
	go func() {
		for range ticker.C {
			c.events <- AliveCellsCount{
				CompletedTurns: response.FinalTurn,
				CellsCount:     len(aliveCellsFromWorld(response.FinalWorld, imageHeight, imageWidth)),
			}
		}
	}()

	finalWorld := makeCall(client, world, p.Turns, p.ImageHeight, p.ImageWidth, c)
	aliveCell := aliveCellsFromWorld(finalWorld, p.ImageHeight, p.ImageWidth)

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          aliveCell,
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
