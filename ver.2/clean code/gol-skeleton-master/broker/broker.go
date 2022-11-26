package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type Broker struct{}

var worker *rpc.Client

func (b *Broker) SendToServer(req stubs.State, res *stubs.State) (err error) {
	aws := flag.String("port", "127.0.0.1:8050", "AWS IP:PORT")
	flag.Parse() // to do : AWS IP:POOR flag 로 받아서 연결
	client, _ := rpc.Dial("tcp", *aws)
	worker = client

	response := new(stubs.State)
	call := worker.Go(stubs.EvaluateAllHandler, stubs.State{World: req.World, Turn: req.Turn}, response, nil)
	fmt.Println("1-1. Broker : Sending Initial world to server ")

	<-call.Done // Get the result of the calculated res.World , res.Turn
	fmt.Println("1-2. Broker : Got computed world from the server ")
	// distributor 의 Ticker 함수에 완료 월드 , 턴 반환 request 로

	res.World = response.World
	res.Turn = response.Turn
	fmt.Println("1-3. Broker: All computed states are ready")

	return
}

func (b *Broker) TickerToServer(req stubs.None, res *stubs.State) (err error) {
	response := new(stubs.State)

	call := worker.Go(stubs.TickerHandler, stubs.None{}, response, nil)
	fmt.Println("2-1. Broker : Sending ticker signal to server ")
	<-call.Done
	fmt.Println("2-2. Broker : Got state for ticker signal ")

	res.World = response.World
	res.Turn = response.Turn

	fmt.Println("2-3. Broker : All ticker states are ready")
	return
}

func (b *Broker) KeyPressToServer(req stubs.KeyPress, res *stubs.State) {

	response := new(stubs.State)
	call := worker.Go(stubs.KeyPressHandler, stubs.KeyPress{KeyPress: req.KeyPress}, response, nil)
	fmt.Println("3-1. Broker : Sending keyPress signal to server ")
	<-call.Done
	fmt.Println("3-2. Broker : Got state for keyPress signal ")

	res.World = response.World
	res.Turn = response.Turn

	fmt.Println("3-3. Broker : All keyPress states are ready")
	return
}

//func (b *Broker) ReturnFinalState(req stubs.Request, res *stubs.Response) {
//	byeWorldTurn(res.ComputedWorld, res.CompletedTurn) }
// server 로 부터 계산 완료된 world 를 (서버 에서 req 로 보냄) 받아 와서 -> distributor 에 전송

func main() {
	listener, _ := net.Listen("tcp", ":8040")
	rpc.Register(&Broker{})
	defer listener.Close()

	rpc.Accept(listener)
}
