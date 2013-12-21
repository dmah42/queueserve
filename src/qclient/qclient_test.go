package qclient

import (
	"testing"
)

const (
	queueName = "testq"
)

var (
	object = []byte("hello queue server")
)

func TestCreate(t *testing.T) {
	id, err := CreateQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	if id != queueName {
		t.Errorf("want %q, got %q", queueName, id)
		return
	}
}

func TestDelete(t *testing.T) {
	err := DeleteQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
}

func TestGet(t *testing.T) {
	_, err := CreateQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected create error: %v", err)
		return
	}
	id, err := GetQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected get error: %v", err)
		return
	}
	if id != queueName {
		t.Errorf("want %q, got %q", queueName, id)
		return
	}
	err = DeleteQueue(id)
	if err != nil {
		t.Errorf("Unexpected delete error: %v", err)
		return
	}
}

func TestEnqueue(t *testing.T) {
	id, err := CreateQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected create error: %v", err)
		return
	}
	err = Enqueue(id, object)
	if err != nil {
		t.Errorf("Unexpected enqueue error: %v", err)
		return
	}
	err = DeleteQueue(queueName)
	if err != nil {
		t.Errorf("Unexpected delete error: %v", err)
		return
	}
}

// TODO: test read
// TODO: concurrent enqueue
// TODO: concurrent dequeue
// TODO: concurrent enqueue/dequeue
