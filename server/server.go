package server

import "uk.ac.bris.cs/gameoflife/util"

func calculateNextAliveCells(p Params, world [][]byte) []util.Cell {
	var aliveCells []util.Cell

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			sum := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					if world[(y+i+p.ImageHeight)%p.ImageHeight][(x+j+p.ImageWidth)%p.ImageWidth] == 0xFF {
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

func worldFromAliveCells(p Params, c []util.Cell) [][]byte {
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}

	for _, i := range c {
		world[i.Y][i.X] = 0xFF
	}

	return world
}

turn := 0
var aliveCells []util.Cell

for y := 0; y < p.ImageHeight; y++ {
for x := 0; x < p.ImageWidth; x++ {
if world[y][x] == 0xFF {
var cell util.Cell
cell.X, cell.Y = x, y
aliveCells = append(aliveCells, cell)
}
}
}

world = worldFromAliveCells(p, aliveCells)


for i := 0; i < p.Turns; i++ {
aliveCells = calculateNextAliveCells(p, world)
world = worldFromAliveCells(p, aliveCells)
turn++
}

func EvaluateAll
