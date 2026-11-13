package websocket

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"taskflow/internal/models"
	"testing"
)

func testClient(id, user, board string, capacity int) *Client {
	return &Client{
		ID:      id,
		UserID:  user,
		BoardID: board,
		Send:    make(chan []byte, capacity),
	}
}

func TestRegisterAndPresence(t *testing.T) {
	hub := NewHub()

	a := testClient("c1", "alice", "board1", 4)
	b := testClient("c2", "bob", "board1", 4)
	c := testClient("c3", "alice", "board1", 4)

	hub.Register(a)
	hub.Register(b)
	hub.Register(c)

	want := []string{"alice", "bob"}
	got := hub.GetPresence("board1")

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("presence = %v, want %v", got, want)
	}

	hub.Unregister(a)

	if got := hub.GetPresence("board1"); !reflect.DeepEqual(got, want) {
		t.Fatalf("presence after first disconnect = %v", got)
	}

	hub.Unregister(c)

	if got := hub.GetPresence("board1"); !reflect.DeepEqual(got, []string{"bob"}) {
		t.Fatalf("presence after second disconnect = %v", got)
	}
}

func TestBroadcastBoardIsolation(t *testing.T) {
	hub := NewHub()

	a := testClient("a", "alice", "board1", 4)
	b := testClient("b", "bob", "board2", 4)

	hub.Register(a)
	hub.Register(b)

	hub.Broadcast(&models.WSEvent{
		Type:    models.EventTaskCreated,
		BoardID: "board1",
		UserID:  "alice",
	})

	select {
	case data := <-a.Send:
		var event models.WSEvent
		if err := json.Unmarshal(data, &event); err != nil {
			t.Fatal(err)
		}
		if event.BoardID != "board1" {
			t.Fatalf("unexpected board: %s", event.BoardID)
		}
	default:
		t.Fatal("board1 client did not receive event")
	}

	select {
	case <-b.Send:
		t.Fatal("board2 client received board1 event")
	default:
	}
}

func TestSlowClientRemoval(t *testing.T) {
	hub := NewHub()
	client := testClient("slow", "alice", "board1", 1)

	hub.Register(client)

	event := &models.WSEvent{
		Type:    models.EventTaskUpdated,
		BoardID: "board1",
	}

	hub.Broadcast(event)
	hub.Broadcast(event)

	if got := hub.GetPresence("board1"); len(got) != 0 {
		t.Fatalf("slow client still present: %v", got)
	}

	// Drain the buffered event and verify the channel was closed.
	_, ok := <-client.Send
	if !ok {
		t.Fatal("expected the first buffered event")
	}

	_, ok = <-client.Send
	if ok {
		t.Fatal("expected closed client channel")
	}

	// Double removal must not panic.
	hub.Unregister(client)
	hub.Unregister(client)
}

func TestReplacementDoesNotRemoveNewClient(t *testing.T) {
	hub := NewHub()

	old := testClient("same-id", "alice", "board1", 2)
	replacement := testClient("same-id", "bob", "board1", 2)

	hub.Register(old)
	hub.Register(replacement)
	hub.Unregister(old)

	got := hub.GetPresence("board1")

	if !reflect.DeepEqual(got, []string{"bob"}) {
		t.Fatalf("replacement unexpectedly removed: %v", got)
	}

	hub.Unregister(replacement)
}

func TestConcurrentConnectionsAndBroadcasts(t *testing.T) {
	hub := NewHub()

	const workers = 32
	const operations = 100

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(worker int) {
			defer wg.Done()

			client := testClient(
				fmt.Sprintf("client-%d", worker),
				fmt.Sprintf("user-%d", worker),
				"board1",
				operations+1,
			)

			hub.Register(client)

			for j := 0; j < operations; j++ {
				hub.Broadcast(&models.WSEvent{
					Type:    models.EventTaskMoved,
					BoardID: "board1",
				})

				_ = hub.GetPresence("board1")
			}

			hub.Unregister(client)
		}(i)
	}

	wg.Wait()

	if users := hub.GetPresence("board1"); len(users) != 0 {
		t.Fatalf("clients leaked after disconnect: %v", users)
	}
}

func TestInvalidEventsAreIgnored(t *testing.T) {
	hub := NewHub()
	client := testClient("c1", "alice", "board1", 4)
	hub.Register(client)

	hub.Broadcast(nil)
	hub.Broadcast(&models.WSEvent{
		Type:    "invalid",
		BoardID: "",
	})

	if got := len(client.Send); got != 0 {
		t.Fatalf("invalid events queued: %d", got)
	}
}
