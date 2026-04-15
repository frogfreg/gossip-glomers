package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"gossip-glomers/handlers"
	"sync"
)

func main() {
	n := maelstrom.NewNode()

	var receivedMessages sync.Map
	var topo handlers.Topology
	var topoChan = make(chan bool)
	var messageChan = make(chan int)

	// var nodeChanMap = make(map[string]chan int)

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	// n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, &topo, nodeChanMap))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, messageChan))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	// n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, nodeChanMap, topoChan))
	n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, topoChan))
	go func() {
		for m := range messageChan {
			for _, neighbor := range topo[n.ID()] {
				body := struct {
					Type    string `json:"type"`
					Message int    `json:"message"`
				}{Type: "broadcast", Message: m}
				go func() {
					// used to send rpc call here
					n.RPC(neighbor)

				}()
			}

		}

	}()

	go func() {
		<-topoChan
		logFile, err := os.Create(fmt.Sprintf("/home/meme/workspace/gossip-glomers/execution-%v.log", n.ID()))
		if err != nil {
			panic(err)
		}

		logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
		slog.Debug("successfully read from topoChan", "topology", topo)

	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}

// func blockingRPC(n *maelstrom.Node, destNode string, message int) error {
// 	var doneChan = make(chan bool)
// 	var errChan = make(chan error)

// 	body := struct {
// 		Type    string `json:"type"`
// 		Message int    `json:"message"`
// 	}{Type: "broadcast", Message: message}

// 	go func() {
// 		slog.Debug("sending rpc", "source", n.ID(), "destination", destNode, "message", message)
// 		err := n.RPC(destNode, body, func(msg maelstrom.Message) error {
// 			slog.Debug("received response from rpc call", "sentFrom", n.ID(), "to", destNode, "message", message)
// 			doneChan <- true
// 			return nil
// 		})
// 		if err != nil {
// 			errChan <- err
// 		} else {
// 			errChan <- nil
// 		}

// 	}()

// 	err := <-errChan
// 	<-doneChan

// 	return err

// }
