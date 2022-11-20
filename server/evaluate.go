package main

import (
	"flag"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// remove for loop
// this function and all of sub-functions will be moved to aws
// all the turns processed on the remote node and get the result back

//gol engine
//responsible for actual processing the turns of game of life
//gol engine as a server on an aws node

//evaluate.go = game of life worker

func calculateNextAliveCells(world [][]byte, imageHeight int, imageWidth int) ([]util.Cell, []util.Cell) {
	var aliveCells []util.Cell
	var flippedCells []util.Cell

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
				} else if sum != 3 {
					flippedCells = append(flippedCells, cell)
				}
			}

			// when a cell has three alive neighbours, it will be alive anyway
			if sum == 3 {
				aliveCells = append(aliveCells, cell)
				if world[y][x] == 0x00 {
					flippedCells = append(flippedCells, cell)
				}
			}
		}
	}
	return aliveCells, flippedCells
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

func (s *GameOfLifeOperation) Evaluate(req stubs.Request, res *stubs.Response) (err error) {
	if req.ShutDown == true {
		os.Exit(0)
	}
	var aliveCells, flippedCells []util.Cell
	world := req.GivenWorld
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

	aliveCells, flippedCells = calculateNextAliveCells(world, imageHeight, imageWidth)
	res.ComputedWorld = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
	res.FlippedCell = flippedCells

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
