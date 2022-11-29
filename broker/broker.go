package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type Broker struct{}

var (
	//workers contains all *rpc.Client for aws nodes
	workers = make([]*rpc.Client, 0)

	//tickerC and keyPressC channels send signal inside to computation for loop
	// and then stateC channel get computing state back
	tickerC   = make(chan bool)
	keyPressC = make(chan rune)
	stateC    = make(chan stubs.State)

	//pause value used for busy waiting
	pause bool
)

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

func (b *Broker) TickerToServer(req stubs.None, res *stubs.State) (err error) {
	tickerC <- true
	currentState := <-stateC
	res.World = currentState.World
	res.Turn = currentState.Turn
	return
}

func (b *Broker) KeyPressToServer(req stubs.KeyPress, res *stubs.State) (err error) {
	keyPressC <- req.KeyPress
	currentState := <-stateC
	res.World = currentState.World
	res.Turn = currentState.Turn
	return
}

func (b *Broker) ShutDown(req stubs.None, res *stubs.None) (err error) {
	none := new(stubs.None)
	for _, w := range workers {
		w.Go(stubs.ShutDownHandler, stubs.None{}, none, nil)
	}
	os.Exit(0)
	return
}

func (b *Broker) SendToServer(req stubs.State, res *stubs.State) (err error) {

	res.World = req.World
	imageHeight := len(res.World)
	imageWidth := len(res.World[0])
	aliveCells := make([]util.Cell, 0)

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if res.World[y][x] == 0xFF {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	for res.Turn = 0; res.Turn < req.Turn; res.Turn++ {
		for pause {
		}
		aliveCellPart := make([]util.Cell, 0)

		go func() {
			for {
				select {
				case <-tickerC:
					stateC <- stubs.State{
						World: res.World,
						Turn:  res.Turn,
					}
				case keyPress := <-keyPressC:
					switch keyPress {
					case 's':
						stateC <- stubs.State{
							World: res.World,
							Turn:  res.Turn,
						}
					case 'q':
						stateC <- stubs.State{
							World: res.World,
							Turn:  res.Turn,
						}
					case 'k':
						stateC <- stubs.State{
							World: res.World,
							Turn:  res.Turn,
						}
						pause = true
					case 'p':
						func() {
							stateC <- stubs.State{
								World: res.World,
								Turn:  res.Turn,
							}
							if pause == false {
								pause = true
							} else if pause == true {
								pause = false
							}
						}()
					}

				}
			}
		}()

		for n, w := range workers {
			aliveCellState := new(stubs.AliveCells)
			w.Call(stubs.EvaluateOneHandler, stubs.EvaluationRequest{
				// TODO : change this to Go
				World:          res.World,
				ID:             n,
				NumberOfWorker: len(workers),
				Turn:           res.Turn,
			}, aliveCellState)
			aliveCellPart = append(aliveCellPart, aliveCellState.AliveCells...)
		}
		aliveCells = aliveCellPart
		if res.Turn == 32 {
			fmt.Println()
		}
		res.World = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
	}
	return
}

func main() {
	server := flag.String("server", "127.0.0.1:8050", "server IP")
	flag.Parse()

	for _, s := range flag.Args() {
		client, _ := rpc.Dial("tcp", *server+":"+s)
		workers = append(workers, client)
	}
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":8040")
	defer listener.Close()
	rpc.Accept(listener)
}
