package main

import (
	"flag"
	"fmt"
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

func calculateNextAliveCells(world [][]byte, imageHeight int, imageWidth int) []util.Cell {
	fmt.Println("calculateNextAliveCells is called")
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

	fmt.Println("calculateNextAliveCells is done")
	return aliveCells
}

func worldFromAliveCells(c []util.Cell, imageHeight int, imageWidth int) [][]byte {
	fmt.Println("worldFromAliveCells is called")
	world := make([][]byte, imageHeight)
	for i := range world {
		world[i] = make([]byte, imageWidth)
	}

	for _, i := range c {
		world[i.Y][i.X] = 0xFF
	}

	fmt.Println("worldFromAliveCells is done")
	return world
}

type GameOfLifeOperation struct{}

func (s *GameOfLifeOperation) EvaluateAll(req stubs.Request, res *stubs.Response) (err error) {
	fmt.Println("let's evaluate the world!!!!!")
	var aliveCells []util.Cell
	world := req.InitialWorld
	turn := req.Turn
	imageHeight := req.ImageHeight
	imageWidth := req.ImageWidth

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if world[y][x] == 0xFF {
				var cell util.Cell
				cell.X, cell.Y = x, y
				aliveCells = append(aliveCells, cell)
			}
		}
	}
	fmt.Println("initial aliveCell is calculated")

	for i := 0; i < turn; i++ {
		aliveCells = calculateNextAliveCells(world, imageHeight, imageWidth)
		world = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
		fmt.Println("turn:", i, "calculated")
	}

	fmt.Println("final world is calculated")
	res.FinalWorld = world
	fmt.Println("response is set")
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
