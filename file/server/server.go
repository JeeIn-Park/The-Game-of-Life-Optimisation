package main

import (
	"flag"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func calculateNextAliveCells(world [][]byte, start int, finish int) []util.Cell {
	aliveCells := make([]util.Cell, 0)
	imageHeight := len(world)
	imageWidth := len(world[0])

	for y := start; y < finish; y++ {
		for x := 0; x < imageHeight; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					//calculate the number of alive neighbour cells including itself
					if world[(y+i+imageHeight)%(imageHeight)][(x+j+imageWidth)%(imageWidth)] == 0xFF {
						sum++
					}
				}
			}
			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
				}
			}
			if sum == 3 {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	return aliveCells
}

type GameOfLifeOperation struct{}

func (g *GameOfLifeOperation) ShutDown(req stubs.None, res *stubs.None) (err error) {
	os.Exit(0)
	return
}

func (g *GameOfLifeOperation) EvaluateOne(req stubs.EvaluationRequest, res *stubs.AliveCells) (err error) {
	imageHeight := len(req.World)
	id := req.ID
	numberOfWorkers := req.NumberOfWorker
	size := (imageHeight - (imageHeight % numberOfWorkers)) / numberOfWorkers
	if id == numberOfWorkers-1 {
		res.AliveCells = calculateNextAliveCells(req.World, id*size, imageHeight)
	} else {
		res.AliveCells = calculateNextAliveCells(req.World, id*size, (id+1)*size)
	}
	return
}

func main() {
	pAddr := flag.String("port", "8000", "Port to listen on")
	//flag.StringVar(&nextAddr, "next", "127.0.0.1:8050", "IP:Port string for next member of the round.")
	flag.Parse()

	rpc.Register(&GameOfLifeOperation{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)

	defer listener.Close()
	rpc.Accept(listener)
}
