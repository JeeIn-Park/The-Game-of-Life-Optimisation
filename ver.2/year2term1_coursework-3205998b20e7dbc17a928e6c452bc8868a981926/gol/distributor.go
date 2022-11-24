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

var dc distributorChannels

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

type GameOfLifeOperation struct{}

func (s *GameOfLifeOperation) Ticker(req stubs.TickerState, res stubs.None) (err error) {
	c.events <- AliveCellsCount{
		CompletedTurns: turn,
		CellsCount:     aliveCellsCount,
	}
	return
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	dc = c
	//server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	//flag.Parse()
	server := "127.0.0.1:8030"
	//client, _ := rpc.Dial("tcp", *server)
	client, _ := rpc.Dial("tcp", server)
	defer client.Close()

	rpc.Register(&GameOfLifeOperation{})

	//io input into world
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

	request := stubs.StartEvaluation{
		InitialWorld: world,
		Turn:         p.Turns,
		ImageHeight:  p.ImageHeight,
		ImageWidth:   p.ImageWidth,
	}
	response := new(stubs.FinishEvaluation)
	done := make(chan *rpc.Call, 10)
	call := client.Go(stubs.EvaluateAllHandler, request, response, done)

	go func() {
		for {
			select {
			//case keyPress := <-keyPresses:
			//switch keyPress {
			//case 's':
			//	fmt.Println("writing pmg image")
			//	writePgm(p, c, world, turn)
			//case 'q':
			//	fmt.Println("q is pressed, quit game of life")
			//	quit(p, c, turn, world, aliveCells, ticker)
			//case 'p':
			//	func() {
			//		if pause == false {
			//			fmt.Println("Paused, current turn is", turn)
			//			pause = true
			//		} else if pause == true {
			//			fmt.Println("Continuing")
			//			pause = false
			//		}
			//	}()
			//
			//}
			case <-purposeC:
			}
		}
	}()

	<-call.Done
	c.events <- FinalTurnComplete{
		CompletedTurns: response.CompletedTrun,
		Alive:          aliveCellFromWorld(p, response.ComputedWorld),
	}

	writePgm(p, c, response.ComputedWorld, response.CompletedTurn)

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}
	close(c.events)
}
