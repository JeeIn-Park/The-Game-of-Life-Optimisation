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

// server.go = game of life worker
var pause bool

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

func (s *GameOfLifeOperation) EvaluateAll(req stubs.Request, res *stubs.Response) (err error) {
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

	//ticker := time.NewTicker(time.Second * 2)
	//go func() {
	//	receive := new(stubs.None)
	//	for range ticker.C {
	//		tickerState := stubs.Response{
	//			ComputedWorld: res.ComputedWorld,
	//			CompletedTurn: res.CompletedTurn,
	//		}
	//		client.Call(stubs.TickerHandler, tickerState, receive)
	//	}
	//}()

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
	flag.Parse()

	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer listener.Close()
	rpc.Accept(listener)
}
