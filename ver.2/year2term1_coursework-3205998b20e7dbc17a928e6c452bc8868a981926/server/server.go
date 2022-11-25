package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// remove for loop
// this function and all of sub-functions will be moved to aws
// all the turns processed on the remote node and get the result back

//gol engine
//responsible for actual processing the turns of game of life
//gol engine as a server on an aws node

//server.go = game of life worker

var (
	nextAddr  string
	pause     bool
	tickerC   = make(chan bool)
	keyPressC = make(chan rune)
	stateC    = make(chan stubs.State)
)

func ticker() {
	fmt.Println("0. ticker is made")
	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		tickerC <- true
		fmt.Println("1. ticker sends the signal")
	}
}

func calculateNextAliveCells(world [][]byte, imageHeight int, imageWidth int) []util.Cell {
	var aliveCells []util.Cell

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					if world[(y+i+imageHeight)%imageHeight][(x+j+imageWidth)%imageWidth] == 0xFF {
						sum++
					}
				}
			}

			var cell util.Cell
			cell.X, cell.Y = x, y

			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, cell)
				}
			}

			if sum == 3 {
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	return aliveCells
}

func worldFromAliveCells(c []util.Cell, imageHeight int, imageWidth int) [][]byte {
	world := make([][]byte, imageHeight)
	for i := range world {
		world[i] = make([]byte, imageWidth)
	}
	for _, i := range c {
		world[i.Y][i.X] = 0xFF
	}
	return world
}

type GameOfLifeOperation struct{}

func (s *GameOfLifeOperation) KeyPress(req stubs.KeyPress, res stubs.State) (err error) {
	keyPressC <- req.KeyPress
	return
}

func (s *GameOfLifeOperation) EvaluateAll(req stubs.InitialInput, res *stubs.State) (err error) {
	var aliveCells []util.Cell
	world := req.InitialWorld
	turn := req.Turn

	imageHeight := len(world)
	imageWidth := len(world[0])
	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	res.CompletedTurn = 0

	go ticker()

	go func() {
		client, _ := rpc.Dial("tcp", nextAddr)
		receive := new(stubs.None)
		for {
			select {
			case <-tickerC:
				fmt.Println("2. get signal from the ticker")
				tickerState := stubs.State{
					ComputedWorld: res.ComputedWorld,
					CompletedTurn: res.CompletedTurn,
				}
				fmt.Println("3. call ticker from the server")
				client.Call(stubs.TickerHandler, tickerState, receive)
				fmt.Println("4. client.Call is successfully done")
			case keyPress := <-keyPressC:
				switch keyPress {
				case 's':
					imageState := stubs.State{
						ComputedWorld: res.ComputedWorld,
						CompletedTurn: res.CompletedTurn,
					}
					client.Call(stubs.KeyPressHandler, imageState, receive)
				case 'q':
					quitState := stubs.State{
						ComputedWorld: res.ComputedWorld,
						CompletedTurn: res.CompletedTurn,
					}
					//이런식으로 콜을 다시 해주는게 아니라 리스폰스를 채널을 통해 전달해서 리스폰스 포인터로 리턴해야할듯
					client.Call(stubs.KeyPressHandler, quitState, receive)
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
		}
	}()

	for i := 0; i < turn; i++ {
		for pause {
		}
		aliveCells = calculateNextAliveCells(world, imageHeight, imageWidth)
		world = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
		res.CompletedTurn++
	}
	res.ComputedWorld = world
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.StringVar(&nextAddr, "next", "127.0.0.1:8050", "IP:Port string for next member of the round.")

	flag.Parse()

	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer listener.Close()
	rpc.Accept(listener)
}
