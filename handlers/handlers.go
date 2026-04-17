package handlers

import (
	"container/list"
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

func BroadcastHandlerFunc(n *maelstrom.Node, receivedMessages *sync.Map, messageChan chan int, topo Topology, nqm map[string]*list.List) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var bb BroadcastBody

		replyBody := BroadcastReply{Type: "broadcast_ok"}

		if err := json.Unmarshal(msg.Body, &bb); err != nil {
			return err
		}

		slog.Debug("broadcast called", "node", n.ID(), "message", bb.Message)
		if _, loaded := receivedMessages.LoadOrStore(bb.Message, true); loaded {
			slog.Debug("already processed", "message", bb.Message)
			return n.Reply(msg, replyBody)
		}

		go func() {
			slog.Debug("sending to messageChan", "message", bb.Message, "current node", n.ID(), "source node", msg.Src)
			messageChan <- bb.Message
		}()

		return n.Reply(msg, replyBody)
	}
}

type ReadReply struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

func ReadHandlerFunc(n *maelstrom.Node, receivedMessages *sync.Map) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {

		slog.Debug("reading local messages", "node", n.ID())

		var finalList []int

		receivedMessages.Range(func(key, value any) bool {
			finalList = append(finalList, key.(int))

			return true
		})

		replyBody := ReadReply{Type: "read_ok", Messages: finalList}

		slog.Debug("reading local messages", "node", n.ID(), "messages", finalList)

		return n.Reply(msg, replyBody)
	}
}

type TopologyReply struct {
	Type string `json:"type"`
}

func TopologyHandlerFunc(n *maelstrom.Node, topo *Topology, topoChan chan bool, nqm map[string]*list.List) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		slog.Debug("topology called", "node", n.ID(), "topology", *topo)

		var tb struct {
			Body Topology `json:"topology"`
		}

		if err := json.Unmarshal(msg.Body, &tb); err != nil {
			return err
		}

		if len(*topo) > 0 {
			return nil
		}
		defer func() {
			topoChan <- true
		}()

		*topo = tb.Body

		for k := range tb.Body {
			if k == n.ID() {

				for _, neighbor := range (*topo)[n.ID()] {

					nqm[neighbor] = list.New()

				}

				continue
			}
			slog.Debug("forwarding topology message", "node", n.ID(), "topology", *topo)
			if err := n.Send(k, msg.Body); err != nil {
				return err
			}
		}

		replyBody := TopologyReply{Type: "topology_ok"}

		slog.Debug("topology should now be set", "node", n.ID(), "local topo", *topo)

		return n.Reply(msg, replyBody)
	}
}
