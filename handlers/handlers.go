package handlers

import (
	"container/list"
	"encoding/json"
	"log/slog"

	"sync"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"time"
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

func BroadcastHandlerFunc(n *maelstrom.Node, receivedMessages *sync.Map, messageChan chan int, topo *Topology, nqm map[string]*list.List) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		slog.Debug("wtf", "topology", topo)
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

		slog.Debug("about to enter neighbor loop ", "neighbors", (*topo)[n.ID()])
		for _, neighbor := range (*topo)[n.ID()] {
			if neighbor == msg.Src {
				continue
			}
			slog.Debug("sending message with retry", "from", n.ID(), "to", neighbor, "message", bb.Message)
			go sendMessageWithRetry(n, neighbor, bb.Message)
		}

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

		topoChan <- true

		*topo = tb.Body

		for _, neighbor := range (*topo)[n.ID()] {
			nqm[neighbor] = list.New()
		}
		topoChan <- true

		replyBody := TopologyReply{Type: "topology_ok"}

		slog.Debug("topology should now be set", "node", n.ID(), "local topo", *topo)

		return n.Reply(msg, replyBody)
	}
}

type broadcastBody struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

func generateBroadcastBody(message int) broadcastBody {
	return broadcastBody{Type: "broadcast", Message: message}
}
func sendMessageWithRetry(n *maelstrom.Node, dest string, message int) {
	done := make(chan bool)
	body := generateBroadcastBody(message)
	for {
		slog.Debug("sending rpc message", "from", n.ID(), "to", dest, "message", message)
		rpcErr := n.RPC(dest, body, func(msg maelstrom.Message) error {
			slog.Debug("received response from node", "from", n.ID(), "to", dest, "msg.src", msg.Src, "msg.dest", msg.Dest, "body", string(msg.Body))
			done <- true

			return nil
		})
		if rpcErr != nil {
			panic(rpcErr)
		}

		select {
		case <-time.After(500 * time.Millisecond):
			slog.Debug("no response, resending message", "source", n.ID(), "destination", dest, "message", message)
			continue
		case <-done:
			slog.Debug("breaking loop in retryable message", "source", n.ID(), "destination", dest, "message", message)
			return

		}
	}

}
