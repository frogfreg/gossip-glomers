package main

import (
	"log"
	"os"

	"gossip-glomers/handlers"
	"sync"

	"container/list"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	var receivedMessages sync.Map
	var topo handlers.Topology
	var topoChan = make(chan bool)

	var nodeChanMap = make(map[string]chan int)

	logFile, err := os.Create("/home/meme/workspace/gossip-glomers/execution.log")
	if err != nil {
		panic(err)
	}

	log.SetOutput(logFile)

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages, &topo, nodeChanMap))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	n.Handle("topology", handlers.TopologyHandlerFunc(n, &topo, nodeChanMap, topoChan))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
