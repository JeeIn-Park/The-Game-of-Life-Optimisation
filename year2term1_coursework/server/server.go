package main

import (
	"flag"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

// 1st define type
//  stubs.go 에서 Exported 된 Type

func calculateNextAliveCells(p gol.Params, turn int, world [][]byte) ([]util.Cell, int) {
	// takes the current state of the world and completes one evolution of the world
	// find next alive cells calculating each cell in the given world
	var aliveCells []util.Cell

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					//calculate the number of alive neighbour cells including itself
					if world[(y+i+p.ImageHeight)%p.ImageHeight][(x+j+p.ImageWidth)%p.ImageWidth] == 0xFF {
						sum++
					}
				}
			}

			var cell util.Cell
			cell.X, cell.Y = x, y

			// when the cell was alive in the given world, exclude it from the number of alive neighbour cells
			// then it keeps alive if it has 2 alive neighbours
			if world[y][x] == 0xFF {
				sum = sum - 1
				if sum == 2 {
					aliveCells = append(aliveCells, cell)
				}
			}

			// when a cell has three alive neighbours, it will be alive anyway
			if sum == 3 {
				aliveCells = append(aliveCells, cell)
			}
		}
	}

	turn++
	return aliveCells, turn
}

type GameOfLifeHandler struct{}

// 얘네는 method

func (s *GameOfLifeHandler) Process(req stubs.Request, res *stubs.Response) (err error) {
	res.AliveCells, res.Turn = calculateNextAliveCells(req.P, req.Turn, req.World)
	return
}

//TODO: for loop 만들기, which iterate over the number of iteration that specified by in the
// request struct

func main() {
	//turn := 0
	//
	//for turn < p.Turns { // p parameter 로 update game of life
	//
	//	updateBoard(c, turn, board, board1, p.ImageHeight, p.ImageWidth, 0, p.ImageHeight) // 0 -> start y
	//
	//	turn = turn + 1
	//} // 이 for 문에 있는 함수, 여기 딸린 sub함수들도 모두 aws 로 가야함.

	pAddr := flag.String("port", "8030", "port..")
	flag.Parse()
	rpc.Register(&GameOfLifeHandler{})

	listner, _ := net.Listen("tcp", ":"+*pAddr)
	defer listner.Close()

	rpc.Accept(listner)
}
