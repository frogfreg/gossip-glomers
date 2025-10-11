package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"gossip-glomers/handlers"
)

func main() {
	n := maelstrom.NewNode()

	var receivedMessages []int

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))
	n.Handle("broadcast", handlers.BroadcastHandlerFunc(n, &receivedMessages))
	n.Handle("read", handlers.ReadHandlerFunc(n, &receivedMessages))
	n.Handle("topology", handlers.TopologyHandlerFunc(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
