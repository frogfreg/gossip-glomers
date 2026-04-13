package handlers

import (
	"encoding/json"
	"log/slog"

	"sync"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Topology map[string][]string

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

func BroadcastHandlerFunc(n *maelstrom.Node, receivedMessages *sync.Map, topo *Topology) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		slog.Info("XXXXXX  broadcast called", "node", n.ID(), "topology", *topo)
		var bb BroadcastBody

		if err := json.Unmarshal(msg.Body, &bb); err != nil {
			return err
		}

		if _, loaded := receivedMessages.LoadOrStore(bb.Message, true); loaded {
			return nil
		}

		for _, neighbor := range (*topo)[n.ID()] {
			if neighbor == msg.Src {
				continue
			}

			if err := n.Send(neighbor, msg.Body); err != nil {
				return err
			}

		}

		replyBody := BroadcastReply{Type: "broadcast_ok"}

		return n.Reply(msg, replyBody)

	}
}

type ReadReply struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

func ReadHandlerFunc(n *maelstrom.Node, receivedMessages *sync.Map) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {

		var finalList []int

		receivedMessages.Range(func(key, value any) bool {
			finalList = append(finalList, key.(int))

			return true
		})

		replyBody := ReadReply{Type: "read_ok", Messages: finalList}

		return n.Reply(msg, replyBody)
	}
}

type TopologyReply struct {
	Type string `json:"type"`
}

func TopologyHandlerFunc(n *maelstrom.Node, topo *Topology) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		slog.Info("XXXXXX  topology called", "node", n.ID(), "topology", *topo)

		var tb struct {
			Body Topology `json:"topology"`
		}

		if err := json.Unmarshal(msg.Body, &tb); err != nil {
			return err
		}

		if len(*topo) > 0 {
			return nil
		}

		*topo = tb.Body

		for k := range tb.Body {
			if k == n.ID() {
				continue
			}
			slog.Info("XXXXXX  forwarding message", "node", n.ID(), "topology", *topo)
			if err := n.Send(k, msg.Body); err != nil {
				return err
			}
		}

		replyBody := TopologyReply{Type: "topology_ok"}

		slog.Info("XXXXXX  topology should now be set", "node", n.ID(), "local topo", *topo)

		return n.Reply(msg, replyBody)
	}
}
