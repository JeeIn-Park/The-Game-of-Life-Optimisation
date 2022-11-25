package stubs

/*ReverseHandler
stubs.ReverseHandler : tell the remote server which method we're calling
ㄴ SecretStingOperation : registered type
ㄴ Reverse : name of the method
request, response are the argument
*/

var EvaluateAllHandler = "GameOfLifeOperation.EvaluateAll"
var KeyPressHandler = "GameOfLifeOperation.KeyPress"

var TickerHandler = "GameOfLifeOperation.Ticker"

/*stubs.go
client uses to call the remote methods on the server
need structure to send request to server and get response back
response struct: a 2d slice (final board state back to local controller)
request struct : a 2d slice (initial state of the board),
				   and other parameters(such as the number of turns, size of image -> so it can iterate the board correctly)
exported method name, exported type (going to be changed to something more appropriate like
									  game of life operations and process turns)
*/

type State struct {
	World [][]byte
	Turn  int
}

type None struct{}

type KeyPress struct {
	KeyPress rune
}
