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

// Reverse // method
// this is method because it's operating on a type(Reverse - exported)
// cannot access reverse string directly with an RPC call
// only able to do it with exported methods
// rename this method more game of life oriented
// have "for loop" here that iterates over the number of iteration specified in the request struct
// once it's done, return it via the response pointer
func (s *GameOfLifeOperation) EvaluateAll(req stubs.Request, res *stubs.Response) (err error) {

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

	for i := 0; i < turn; i++ {
		aliveCells = calculateNextAliveCells(world, imageHeight, imageWidth)
		world = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
		turn++
	}

	res.FinalWorld = world
	return
}

func main() {
	//concerned with just getting the port that we're going to listen on
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()

	//kind of boilerplate code for registering "SecretStringOperations"type
	//when you register this type, it's exported methods will be able to be called remotely
	//look at the client
	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	//closes it when everything's done
	defer listener.Close()
	rpc.Accept(listener)
}
