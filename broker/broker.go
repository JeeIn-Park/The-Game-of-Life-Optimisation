package main

import (
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type Broker struct{}

var workers = make([]*rpc.Client, 0)

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

func (b *Broker) SendToServer(req stubs.State, res *stubs.State) (err error) {

	res.World = req.World
	imageHeight := len(res.World)
	imageWidth := len(res.World[0])
	aliveCells := make([]util.Cell, 0)
	aliveCellState := new(stubs.AliveCells)

	for y := 0; y < imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			if res.World[y][x] == 0xFF {
				aliveCells = append(aliveCells, util.Cell{X: x, Y: y})
			}
		}
	}

	for res.Turn = 0; res.Turn < req.Turn; res.Turn++ {
		aliveCellPart := make([]util.Cell, 0)
		for n, w := range workers {
			w.Call(stubs.EvaluateOneHandler, stubs.EvaluationRequest{
				World:          res.World,
				ID:             n,
				NumberOfWorker: len(workers),
			}, aliveCellState)
			aliveCellPart = append(aliveCellPart, aliveCellState.AliveCells...)
		}
		aliveCells = aliveCellPart
		res.World = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
	}
	return
}

//func (b *Broker) TickerToServer(req stubs.None, res *stubs.State) (err error) {
//	response := new(stubs.State)
//	call := worker.Go(stubs.TickerHandler, stubs.None{}, response, nil)
//	<-call.Done
//	res.World = response.World
//	res.Turn = response.Turn
//	return
//}

//func (b *Broker) KeyPressToServer(req stubs.KeyPress, res *stubs.State) (err error) {
//	response := new(stubs.State)
//	call := worker.Go(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: req.KeyPress}, response, nil)
//	fmt.Println("3-1. Broker : Sending keyPress signal to server ")
//	<-call.Done
//	fmt.Println("3-2. Broker : Got state for keyPress signal ")
//
//	res.World = response.World
//	res.Turn = response.Turn
//
//	fmt.Println("3-3. Broker : All keyPress states are ready")
//	return
//}

//func (b *Broker) ShutDown(req stubs.None, res stubs.None) (err error) {
//	worker.Go(stubs.ShutDownHandler, stubs.None{}, stubs.None{}, nil)
//	os.Exit(0)
//	return
//}

//func (b *Broker) ReturnFinalState(req stubs.Request, res *stubs.Response) {
//	byeWorldTurn(res.ComputedWorld, res.CompletedTurn) }
// server 로 부터 계산 완료된 world 를 (서버 에서 req 로 보냄) 받아 와서 -> distributor 에 전송

func main() {
	listener, _ := net.Listen("tcp", ":8040")
	rpc.Register(&Broker{})
	defer listener.Close()

	client, _ := rpc.Dial("tcp", "127.0.0.1:8050")
	workers = append(workers, client)
	//TODO : here needs to be modified to be able to accept multiple servers

	rpc.Accept(listener)
}
