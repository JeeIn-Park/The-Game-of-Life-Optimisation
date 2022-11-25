package gol

import (
	"fmt"
	"net"
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

var dc distributorChannels

func aliveCellFromWorld(world [][]byte, imageHeight int, imageWidth int) []util.Cell {
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

func writePgm(world [][]byte, turn int, imageHeight int, imageWidth int) {
	dc.ioCommand <- ioOutput
	dc.ioFilename <- fmt.Sprintf("%dx%dx%d", imageHeight, imageWidth, turn)
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			dc.ioOutput <- world[y][x]
		}
	}
}

func quit(c distributorChannels, turn int, world [][]byte) {
	imageHeight := len(world)
	imageWidth := len(world[0])

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCellFromWorld(world, imageHeight, imageWidth),
	}
	writePgm(world, turn, imageHeight, imageWidth)
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	close(c.events)

}

type GameOfLifeOperation struct{}

func (s *GameOfLifeOperation) Ticker(req stubs.State, res *stubs.None) (err error) {
	dc.events <- AliveCellsCount{
		CompletedTurns: req.CompletedTurn,
		CellsCount:     len(aliveCellFromWorld(req.ComputedWorld, len(req.ComputedWorld), len(req.ComputedWorld[0]))),
	}
	return
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	dc = c

	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	//flag.Parse()
	// pAddr := flag.String("port", "8050", "Port to listen on")
	server := "127.0.0.1:8030"
	//client, _ := rpc.Dial("tcp", *server)
	client, _ := rpc.Dial("tcp", server)
	listener, _ := net.Listen("tcp", ":8050")
	defer listener.Close()
	defer client.Close()
	rpc.Register(&GameOfLifeOperation{})
	go rpc.Accept(listener)

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

	request := stubs.InitialInput{
		InitialWorld: world,
		Turn:         p.Turns,
	}
	response := new(stubs.State)
	call := client.Go(stubs.EvaluateAllHandler, request, response, nil)
	pause := false

	go func() {
		for {
			keyPress := <-keyPresses
			switch keyPress {
			case 's':
				fmt.Println("writing pmg image")
				client.Call(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: keyPress}, response)
				writePgm(response.ComputedWorld, response.CompletedTurn, p.ImageHeight, p.ImageWidth)
			case 'q':
				fmt.Println("q is pressed, quit game of life")
				client.Call(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: keyPress}, response)
				quit(c, response.CompletedTurn, response.ComputedWorld)
			case 'k':
				fmt.Println("k is pressed, shutting down")
				err := client.Call(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: keyPress}, response)
				if err != nil {
					fmt.Println()
				}
				quit(c, response.CompletedTurn, response.ComputedWorld)
			case 'p':
				func() {
					if pause == false {
						fmt.Println("p is pressed, pausing")
						client.Call(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: keyPress}, response)
						fmt.Println("Paused, current turn is", response.CompletedTurn)
						pause = true
					} else if pause == true {
						fmt.Println("p is pressed, continuing")
						client.Call(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: keyPress}, response)
						fmt.Println("Continuing")
						pause = false
					}
				}()
			}
		}

	}()

	<-call.Done
	quit(c, response.CompletedTurn, response.ComputedWorld)
}

/*
func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	flag.Parse()
	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
*/
