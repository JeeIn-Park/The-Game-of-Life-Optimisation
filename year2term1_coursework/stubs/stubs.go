package stubs

import (
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

var GameOfLifeHandler = "GameOfLifeHandler.Process"

type Response struct {
	AliveCells []util.Cell
	Turn       int
}

type Request struct {
	P     gol.Params
	Turn  int
	World [][]byte
}
