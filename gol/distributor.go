package gol

import (
	"bufio"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

//parameter: rpc.Client
//needed when make remote procedure call
//parameter: message string
//need to replace with 2d slice and some parameters

// func makeCall(client rpc.Client, message string)
func makeCall(client *rpc.Client, message string) {

	request := stubs.Request{Message: message}
	//use new for this
	//so this going to be a pointer
	//it is important for "client.call"
	response := new(stubs.Response)
	//when you make remote procedure call,
	//the response argument needs to be a pointer to itself
	//stubs.ReverseHandler : tell the remote server which method we're calling
	//ㄴ SecretStingOperation : registered type
	//ㄴ Reverse : name of the method
	//request, response are the argument
	client.Call(stubs.ReverseHandler, request, response)
	fmt.Println("Responded: " + response.Message)
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	//local controller
	//responsible for io and capturing keypress
	//local controller as a client on a local machine

	//client.go = local controller

	//step 1
	//implementing a basic controller
	//which can tell the logic engine to evolve game of life
	//for the number of turns specified in gol.Params.Turns
	//** can achieve this by implementing a single, blocking RPC call to process all requested turns

	//stubs.go
	//client uses to call the remote methods on the server
	//need structure to send request to server and get response back
	//response struct: a 2d slice (final board state back to local controller)
	//request struct : a 2d slice (initial state of the board),
	//				   and other parameters(such as the number of turns, size of image -> so it can iterate the board correctly)
	//exported method name, exported type (going to be changed to something more appropriate like
	//									  game of life operations and process turns)

	//loading the board from io
	//keep this on the client side
	world := make([][]byte, p.ImageHeight)
	for i := range world {
		world[i] = make([]byte, p.ImageWidth)
	}
	c.ioCommand <- ioInput
	c.ioFilename <- fmt.Sprintf("%dx%d", p.ImageHeight, p.ImageWidth)
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			world[y][x] = <-c.ioInput
		}
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

	// remove for loop
	// this function and all of sub-functions will be moved to aws
	// all the turns processed on the remote node and get the result back

	//gol engine
	//responsible for actual processing the turns of game of life
	//gol engine as a server on an aws node

	//server.go = game of life worker

	c.events <- FinalTurnComplete{
		CompletedTurns: turn,
		Alive:          aliveCells,
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	close(c.events)
}

func main() {
	//it's fine when you practice with localhost
	//actually need to use aws node
	//do not just use local host for the actual submission
	//server := "127.0.0.1:8030"

	//boilerplate code again
	//again refers to the golang documentation,
	//you can directly copy this into assignment
	//client, err := rpc.Dial("tcp", sever)
	/*
		if err != nil {
			log.Fatal("dialing:", err)
		}
		defer client.Close()
	*/

	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	file, _ := os.Open("wordlist")
	scanner := bufio.NewScanner(file)
	//this loop iterating over a text file and sending strings to the server
	// so that's not the same as the game of life
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Println("Called: " + t)
		//makeCall(*client, t)
		//pass the copu of it to makeCall function
		// client stucture contains mutex lock
		// copying mutex lock is a problem(two train allowed)
		makeCall(client, t)
	}

}
