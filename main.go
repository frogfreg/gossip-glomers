package main

import (
	"log"
	"os"

	"gossip-glomers/handlers"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	var receivedMessages sync.Map
	var topo handlers.Topology

	logFile, err := os.Create("/home/nyan/temp/maelstrom/logs/execution.log")
	if err != nil {
		panic(err)
	}

	log.SetOutput(logFile)

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, &topo))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
