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

func writePgm(c distributorChannels, world [][]byte, turn int, imageHeight int, imageWidth int) {
	c.ioCommand <- ioOutput
	c.ioFilename <- fmt.Sprintf("%dx%dx%d", imageHeight, imageWidth, turn)
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			c.ioOutput <- world[y][x]
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
	writePgm(c, world, turn, imageHeight, imageWidth)
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	close(c.events)

}

type GameOfLifeOperation struct{}

func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	client, _ := rpc.Dial("tcp", "127.0.0.1:8040")

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

	request := stubs.State{
		World: world,
		Turn:  p.Turns,
	}
	response := new(stubs.State)
	call := client.Go(stubs.SendToServer, request, response, nil)
	pause := false

	ticker := time.NewTicker(time.Second * 2)
	go func() {
		response := new(stubs.State)
		for range ticker.C {
			if pause == false {
				client.Call(stubs.TickerToServer, stubs.None{}, response)
				c.events <- AliveCellsCount{
					CompletedTurns: response.Turn,
					CellsCount:     len(aliveCellFromWorld(response.World, p.ImageHeight, p.ImageWidth)),
				}
			}
		}
	}()

	go func() {
		for {
			keyPress := <-keyPresses
			switch keyPress {
			case 's':
				response := new(stubs.State)
				fmt.Println("writing pmg image")
				client.Call(stubs.KeyPressToServer, stubs.KeyPress{KeyPress: keyPress}, response)
				writePgm(c, response.World, response.Turn, p.ImageHeight, p.ImageWidth)
			case 'q':
				response := new(stubs.State)
				fmt.Println("q is pressed, quit game of life")
				client.Call(stubs.KeyPressToServer, stubs.KeyPress{KeyPress: keyPress}, response)
				quit(c, response.Turn, response.World)
			case 'k':
				response := new(stubs.State)
				fmt.Println("k is pressed, shutting down")
				pause = true
				client.Call(stubs.KeyPressToServer, stubs.KeyPress{KeyPress: keyPress}, response)
				none := new(stubs.None)
				client.Go(stubs.ShutDown, stubs.None{}, none, nil)
				quit(c, response.Turn, response.World)
			case 'p':
				func() {
					response := new(stubs.State)
					if pause == false {
						fmt.Println("p is pressed, pausing")
						client.Call(stubs.KeyPressToServer, stubs.KeyPress{KeyPress: keyPress}, response)
						fmt.Println("Paused, current turn is", response.Turn)
						pause = true
					} else if pause == true {
						fmt.Println("p is pressed, continuing")
						client.Call(stubs.KeyPressToServer, stubs.KeyPress{KeyPress: keyPress}, response)
						fmt.Println("Continuing")
						pause = false
					}
				}()
			}
		}
	}()

	<-call.Done
	quit(c, response.Turn, response.World)
}
