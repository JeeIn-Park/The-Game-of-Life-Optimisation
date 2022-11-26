package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
)

type Broker struct{}

func (b *Broker) HiWorld(req stubs.InitialInput, res *stubs.State) (err error) {
	response := new(stubs.State)
	aws := "127.0.0.1:8050" // to do : AWS IP:POOR flag 로 받아서 연결
	client, _ := rpc.Dial("tcp", aws)

	call := client.Go(stubs.EvaluateAllHandler, stubs.InitialInput{InitialWorld: req.InitialWorld, Turn: req.Turn}, response, nil)
	fmt.Println("1. Broker : Sending Initial world to server ")

	<-call.Done // Get the result of the calculated res.World , res.Turn
	fmt.Println("2. Broker : Got computed world from the server ")

	res.ComputedWorld = response.ComputedWorld
	res.CompletedTurn = response.CompletedTurn

	fmt.Println("3. Broker: All computed states are ready")

	// rpc.Dial("tcp", "127.0.0.1:8030") 없어도 작동?

	return
}

//func (b *Broker) ReturnFinalState(req stubs.Request, res *stubs.Response) {
//	byeWorldTurn(res.ComputedWorld, res.CompletedTurn) }
// server 로 부터 계산 완료된 world 를 (서버 에서 req 로 보냄) 받아 와서 -> distributor 에 전송

func main() {

	pAddr := flag.String("port", "8040", "Broker's Port to listen on")
	flag.Parse()
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)

}
