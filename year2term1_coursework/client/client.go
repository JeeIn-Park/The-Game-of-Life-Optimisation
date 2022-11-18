package main

import (
	"flag"
	"fmt"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

func main() {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	fmt.Println("Server: ", *server)
	//TODO: connect to the RPC server and send the request(s)
	client, _ := rpc.Dial("tcp", *server)


	for i := 0; i < p.Turns; i++ {
		request := stubs.Request{
			P:     p,
			World: world,
		}
		response := new(stubs.Response)

		client.Call(stubs.GameOfLifeHandler, request, response)
		// aliveCells = response.AliveCells
		responseCellOut := make(chan []util.Cell)
		responseTurnOut := make(chan int)
		responseCellOut <- response.AliveCells
		responseTurnOut <- response.Turn

		defer client.Close()
}
