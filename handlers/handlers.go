package handlers

import (
	"encoding/json"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func EchoHandlerFunc(n *maelstrom.Node) func(maelstrom.Message) error {

	return func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "echo_ok"

		return n.Reply(msg, body)

	}
}

type GenerateReply struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

func GenerateHandlerFunc(n *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		newId := uuid.NewString()

		replyBody := GenerateReply{Type: "generate_ok", Id: newId}

		return n.Reply(msg, replyBody)

	}
}

type BroadcastBody struct {
	Message int `json:"message"`
}

type BroadcastReply struct {
	Type string `json:"type"`
}

func BroadcastHandlerFunc(n *maelstrom.Node, receivedMessages *[]int) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var bb BroadcastBody

		if err := json.Unmarshal(msg.Body, &bb); err != nil {
			return err
		}

		*receivedMessages = append(*receivedMessages, bb.Message)

		replyBody := BroadcastReply{Type: "broadcast_ok"}

		return n.Reply(msg, replyBody)

	}
}

type ReadReply struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

func ReadHandlerFunc(n *maelstrom.Node, receivedMessages *[]int) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {

		replyBody := ReadReply{Type: "read_ok", Messages: *receivedMessages}

		return n.Reply(msg, replyBody)
	}
}

type TopologyReply struct {
	Type string `json:"type"`
}

func TopologyHandlerFunc(n *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {

		replyBody := TopologyReply{Type: "topology_ok"}

		return n.Reply(msg, replyBody)
	}
}
