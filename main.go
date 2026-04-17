package main

import (
	"container/list"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"gossip-glomers/handlers"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	var receivedMessages sync.Map
	var topo handlers.Topology
	var topoChan = make(chan bool)
	var messageChan = make(chan int)
	var nodeQueueMap = make(map[string]*list.List)

	// var nodeChanMap = make(map[string]chan int)

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	// n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, &topo, nodeChanMap))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, messageChan, topo, nodeQueueMap))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	// n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, nodeChanMap, topoChan))
	n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, topoChan, nodeQueueMap))
	go func() {
		for m := range messageChan {
			for _, neighbor := range topo[n.ID()] {
				l := nodeQueueMap[neighbor]
				slog.Debug("pushing to queue", "node queue", neighbor, "message", m)
				l.PushBack(m)
				var stop = false
				for e := l.Front(); e != nil && !stop; e = e.Next() {
					var done = make(chan bool)
					body := generateBroadcastBody((e.Value).(int))

					slog.Debug("about to send rpc message", "message", e.Value, "from", n.ID(), "to", neighbor)
					rpcErr := n.RPC(neighbor, body, func(msg maelstrom.Message) error {
						slog.Debug("received response from node", "from", n.ID(), "to", neighbor, "msg.src", msg.Src, "msg.dest", msg.Dest, "body", string(msg.Body))
						done <- true

						return nil
					})
					if rpcErr != nil {
						panic(rpcErr)
					}

					select {
					case <-time.After(900 * time.Millisecond):
						slog.Debug("no response, stopping loop", "message", e.Value, "from", n.ID(), "to", neighbor, "queue length", l.Len())

						stop = true
					case <-done:
						slog.Debug("successfully processed element from queue", "message", e.Value)
						l.Remove(e)

					}
				}
				slog.Debug("finished processing neighbor", "neighbor", neighbor)
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

type broadcastBody struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

func generateBroadcastBody(message int) broadcastBody {
	return broadcastBody{Type: "broadcast", Message: message}
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

// body := struct {
// 	Type    string `json:"type"`
// 	Message int    `json:"message"`
// }{Type: "broadcast", Message: m}
