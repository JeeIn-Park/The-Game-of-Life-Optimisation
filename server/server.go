package main

import (
	"flag"
	"net"
	"net/rpc"
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
	//	nextAddr  string
	pause     bool
	tickerC   = make(chan bool)
	keyPressC = make(chan rune)
	stateC    = make(chan stubs.State)
)

func calculateNextAliveCells(world [][]byte, start int, finish int) []util.Cell {
	var aliveCells []util.Cell
	imageHeight := len(world)
	imageWidth := len(world[0])

	for y := start; y < finish; y++ {
		for x := 0; x < imageHeight; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					if world[(y+i+imageHeight)%imageHeight][(x+j+imageWidth)%imageWidth] == 0xFF {
						sum++
					}
				}
			}

			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
				}

				if sum == 3 {
					aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
				}
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

//func (g *GameOfLifeOperation) KeyPress(req stubs.KeyPress, res *stubs.State) (err error) {
//	fmt.Println("Got keyPress from distributor.go correctly")
//	keyPressC <- req.KeyPress
//	currentState := <-stateC
//	res.World = currentState.World
//	res.Turn = currentState.Turn
//	fmt.Println("All states are registered correctly")
//	return
//}
//
//func (g *GameOfLifeOperation) Ticker(req stubs.None, res *stubs.State) (err error) {
//	tickerC <- true
//	currentState := <-stateC
//	res.World = currentState.World
//	res.Turn = currentState.Turn
//	return
//}
//
//func (g *GameOfLifeOperation) ShutDown(req stubs.None, res *stubs.None) (err error) {
//	os.Exit(0)
//	return
//}

func (g *GameOfLifeOperation) EvaluateOne(req stubs.EvaluationRequest, res *stubs.AliveCellState) (err error) {
	imageHeight := len(req.World)
	id := req.ID
	numberOfWorkers := req.NumberOfWorker

	size := (imageHeight - (imageHeight % numberOfWorkers)) / numberOfWorkers
	//remove possibility of remainder
	//remained world parts will be calculated at the last part

	var aliveCells []util.Cell
	if id == numberOfWorkers-1 {
		aliveCells = calculateNextAliveCells(req.World, id*size, imageHeight)
	} else {
		aliveCells = calculateNextAliveCells(req.World, id*size, (id+1)*size)
	}
	//for y := 0; y < imageHeight; y++ {
	//	for x := 0; x < imageWidth; x++ {
	//		if req.World[y][x] == 0xFF {
	//			var cell util.Cell
	//			cell.X, cell.Y = x, y
	//			aliveCells = append(aliveCells, cell)
	//		}
	//	}
	//

	// aliveCells = calculateNextAliveCells(req.World, imageHeight, imageWidth)
	res.AliveCells = aliveCells
	return
}

//func (s *GameOfLifeOperation) EvaluateAll(req stubs.State, res *stubs.State) (err error) {
//	var aliveCells []util.Cell
//	res.World = req.World
//	turn := req.Turn
//
//	imageHeight := len(res.World)
//	imageWidth := len(res.World[0])
//	for y := 0; y < imageHeight; y++ {
//		for x := 0; x < imageWidth; x++ {
//			if res.World[y][x] == 0xFF {
//				var cell util.Cell
//				cell.X, cell.Y = x, y
//				aliveCells = append(aliveCells, cell)
//			}
//		}
//	}
//	res.Turn = 0
//
//	go func() {
//		for {
//			select {
//			case <-tickerC:
//				stateC <- stubs.State{
//					World: res.World,
//					Turn:  res.Turn,
//				}
//			case keyPress := <-keyPressC:
//				switch keyPress {
//				case 's':
//					stateC <- stubs.State{
//						World: res.World,
//						Turn:  res.Turn,
//					}
//				case 'q':
//					stateC <- stubs.State{
//						World: res.World,
//						Turn:  res.Turn,
//					}
//				case 'k':
//					stateC <- stubs.State{
//						World: res.World,
//						Turn:  res.Turn,
//					}
//					pause = true
//				case 'p':
//					func() {
//						stateC <- stubs.State{
//							World: res.World,
//							Turn:  res.Turn,
//						}
//						if pause == false {
//							pause = true
//						} else if pause == true {
//							pause = false
//						}
//					}()
//				}
//
//			}
//		}
//	}()
//
//	for i := 0; i < turn; i++ {
//		for pause {
//		}
//		aliveCells = calculateNextAliveCells(res.World, imageHeight, imageWidth)
//		res.World = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
//		res.Turn++
//	}
//
//	return
//}

func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	//flag.StringVar(&nextAddr, "next", "127.0.0.1:8050", "IP:Port string for next member of the round.")
	flag.Parse()

	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer listener.Close()
	rpc.Accept(listener)
}
