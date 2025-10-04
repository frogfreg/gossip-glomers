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

		newBody, err := json.Marshal(GenerateReply{Type: "generate_ok", Id: newId})
		if err != nil {
			return err
		}

		return n.Reply(msg, newBody)

	}
}
