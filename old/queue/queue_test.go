package queue

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	q := New()
	if len(q.items) != 0 {
		t.Errorf("Expected queue to be empty but it has %d items", len(q.items))
	}
}

func TestPush(t *testing.T) {
	q := New()
	for i := 0; i < 10; i++ {
		q.Push([]byte{0x01, 0x02, 0x03, 0x04})
	}
	if len(q.items) != 10 {
		t.Errorf("Pushed 10 items onto the queue but it only has %d items", len(q.items))
	}
}

func TestPop(t *testing.T) {
	q := New()
	items := [][]byte{{0x00}, {0x11}, {0x22}, {0x33}, {0x44}}
	for _, item := range items {
		q.Push(item)
	}
	for i := 0; i < len(items); i++ {
		popped := q.Pop()
		if bytes.Compare(popped, items[i]) != 0 {
			t.Errorf("Popped %v but expected %v", popped, items[i])
		}
	}
}
