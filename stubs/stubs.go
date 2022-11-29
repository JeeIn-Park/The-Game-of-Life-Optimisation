package stubs

import "uk.ac.bris.cs/gameoflife/util"

/*stubs.go
client uses to call the remote methods on the server
need structure to send request to server and get response back
*/

var ShutDownHandler = "GameOfLifeOperation.ShutDown"
var EvaluateOneHandler = "GameOfLifeOperation.EvaluateOne"

var SendToServer = "Broker.SendToServer"
var TickerToServer = "Broker.TickerToServer"
var KeyPressToServer = "Broker.KeyPressToServer"
var ShutDown = "Broker.ShutDown"

type State struct {
	World [][]byte
	Turn  int
}

type None struct{}

type KeyPress struct {
	KeyPress rune
}

type AliveCells struct {
	AliveCells []util.Cell
}

type EvaluationRequest struct {
	World          [][]byte
	ID             int
	NumberOfWorker int
	Turn           int
}
