package main

import (
	"flag"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
)

var (
	//topics  = make(map[string]chan stubs.Pair)
	//topicmx sync.RWMutex

	//workers []*rpc.Client
	worker rpc.Client
)

// Create a new topic as a buffered channel.
//func createTopic(topic string, buflen int) {
//	topicmx.Lock()
//	defer topicmx.Unlock()
//	if _, ok := topics[topic]; !ok {
//		topics[topic] = make(chan stubs.Pair, buflen)
//		fmt.Println("Created channel #", topic)
//	}
//}

//func addWorker(workerAddress string){
//	worker, _ := rpc.Dial("tcp", workerAddress)
//	workers = append(workers, worker)
//}

// The Pair is published to the topic.
//func publish(topic string, pair stubs.Pair) (err error) {
//	topicmx.RLock()
//	defer topicmx.RUnlock()
//	if ch, ok := topics[topic]; ok {
//		ch <- pair
//	} else {
//		return errors.New("No such topic.")
//	}
//	return
//}

// The subscriber loops run asynchronously, reading from the topic and sending the err
// 'job' pairs to their associated subscriber.
//func subscriber_loop(topic chan stubs.Pair, client *rpc.Client, callback string) {
//	for {
//		job := <-topic
//		response := new(stubs.JobReport)
//		err := client.Call(callback, job, response)
//		if err != nil {
//			fmt.Println("Error")
//			fmt.Println(err)
//			fmt.Println("Closing subscriber thread.")
//			//Place the unfulfilled job back on the topic channel.
//			topic <- job
//			break
//		}
//		fmt.Println(callback, "of", job.X, "and", job.Y, "is", response.Result)
//	}
//}

// The subscribe function registers a worker to the topic, creating an RPC client,
// and will use the given callback string as the callback function whenever work
// is available.
//func subscribe(topic string, factoryAddress string, callback string) (err error) {
//	fmt.Println("Subscription request")
//	topicmx.RLock()
//	ch := topics[topic]
//	topicmx.RUnlock()
//	client, err := rpc.Dial("tcp", factoryAddress)
//	if err == nil {
//		go subscriber_loop(ch, client, callback)
//	} else {
//		fmt.Println("Error subscribing ", factoryAddress)
//		fmt.Println(err)
//		return err
//	}
//	return
//}

type Broker struct{}

//func (b *Broker) CreateChannel(req stubs.ChannelRequest, res *stubs.StatusReport) (err error) {
//	createTopic(req.Topic, req.Buffer)
//	return
//}

//func (b *Broker) Subscribe(req stubs.Subscription, res *stubs.StatusReport) (err error) {
//	err = subscribe(req.Topic, req.FactoryAddress, req.Callback)
//	if err != nil {
//		res.Message = "Error during subscription"
//	}
//	return err
//}

func (b *Broker) Initialise(req stubs.State, res *stubs.State) (err error) {
	//err = initialise(req.World, req.Turn)
	//return err
	//for i , worker := range workers{
	//	finalWorld[i] := new(stubs.State)
	//	worker.Go(stubs.EvaluateAllHandler, stubs.State{ World: req.World, Turn:  req.Turn}, finalWorld, nil)
	//}

	worker.Go(stubs.EvaluateAllHandler, stubs.State{World: req.World, Turn: req.Turn}, finalWorld, nil)

	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&Broker{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}
