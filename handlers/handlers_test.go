package handlers

import (
	"testing"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func TestStuff(t *testing.T) {
	n := maelstrom.NewNode()
	topo := make(Topology)
	var topoChan = make(chan bool)
	go func() {
		<-topoChan
	}()

	f := TopologyHandlerFunc(n, &topo, topoChan)

	m := maelstrom.Message{
		Src:  "c7",
		Dest: "n0",
		Body: []byte(`{"type":"topology","topology":{"n0":["n3","n1"],"n1":["n4","n2","n0"],"n2":["n1"],"n3":["n0","n4"],"n4":["n1","n3"]},"msg_id":1}`),
	}

	t.Logf("topology before being overwritten: %+v\n", topo)

	if err := f(m); err != nil {
		t.Error(err)
	}

	// t.Logf("topology after calling handler: %+v\n", topo)
	// t.Error("erring on purpose")

}
