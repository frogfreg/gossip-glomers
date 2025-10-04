package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"gossip-glomers/handlers"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", handlers.EchoHandlerFunc(n))
	n.Handle("generate", handlers.GenerateHandlerFunc(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
