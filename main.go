package main

import (
	"container/list"
	"fmt"
	"log"
	"log/slog"
	"os"

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

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, messageChan, &topo, nodeQueueMap))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, topoChan, nodeQueueMap))

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
