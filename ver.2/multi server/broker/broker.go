package main

import (
	"flag"
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
	//for i := 0; i < len(req.Ports); i ++ {
	//	client, _ := rpc.Dial("tcp", req.InstanceIP+":"+req.Ports[i])
	//	workers = append(workers, client)
	//}

	aws := flag.String("port", "127.0.0.1:8050", "AWS IP:PORT")
	flag.Parse() // to do : AWS IP:POOR flag 로 받아서 연결
	client, _ := rpc.Dial("tcp", *aws)
	workers = append(workers, client)
	//TODO : here needs to be modified to be able to accept multiple servers

	//worker = client
	//
	//response := new(stubs.State)
	//call := worker.Go(stubs.EvaluateAllHandler, stubs.State{World: req.World, Turn: req.Turn}, response, nil)
	//<-call.Done // Get the result of the calculated res.World , res.Turn
	//// distributor 의 Ticker 함수에 완료 월드 , 턴 반환 request 로
	//res.World = response.World
	//res.Turn = response.Turn

	res.World = req.World
	computingTurn := 0
	imageHeight := len(res.World)
	imageWidth := len(res.World[0])
	aliveCellState := make([]stubs.AliveCellState, len(workers))
	calls := make([]*rpc.Call, len(workers))
	aliveCells := make([]util.Cell, 0)
	for i := 0; i < req.Turn; i++ {
		for n, w := range workers {
			calls[n] = w.Go(stubs.EvaluateOneHandler, stubs.EvaluationRequest{
				World:          res.World,
				Turn:           computingTurn,
				ID:             n,
				NumberOfWorker: len(workers),
			}, aliveCellState[n], nil)
		}
		for n := range workers {
			<-calls[n].Done
			aliveCells = append(aliveCells, aliveCellState[n].AliveCells...)
		}
		res.World = worldFromAliveCells(aliveCells, imageHeight, imageWidth)
		computingTurn++
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

	rpc.Accept(listener)
}
