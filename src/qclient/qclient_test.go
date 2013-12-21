package qclient

import (
	"bytes"
	"flag"
	"testing"
	"time"
)

const (
	queueName = "testq"
)

var (
	object = []byte("hello queue server")
	host = flag.String("host", "localhost", "qserver host")
	port = flag.Int("port", 4242, "qserver port")
)

func initQClient() {
	flag.Parse()
	Host = *host
	Port = *port
}

func TestCreate(t *testing.T) {
	initQClient()
	id, err := CreateQueue(queueName)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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
		t.Errorf("unexpected error: %v", err)
		return
	}
}

func TestGet(t *testing.T) {
	CreateQueue(queueName)
	id, err := GetQueue(queueName)
	if err != nil {
		t.Errorf("unexpected get error: %v", err)
		return
	}
	if id != queueName {
		t.Errorf("want %q, got %q", queueName, id)
		return
	}
	DeleteQueue(id)
}

func TestEnqueue(t *testing.T) {
	id, _ := CreateQueue(queueName)
	err := Enqueue(id, object)
	if err != nil {
		t.Errorf("unexpected enqueue error: %v", err)
		return
	}
	DeleteQueue(queueName)
}

func TestReadDequeue(t *testing.T) {
	id, _ := CreateQueue(queueName)
	Enqueue(id, object)
	response, err := Read(id, 2)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
		return
	}
	if response.Id != id || !bytes.Equal(response.Object, object) {
		t.Errorf("want %q | %q, got %q | %q", response.Id, response.Object, id, object)
		return
	}
	if err := Dequeue(id, response.EntityId); err != nil {
		t.Errorf("unexpected dequeue error: %v", err)
		return
	}

	_, err = Read(id, 2)
	if err == nil {
		t.Errorf("expected read error on empty queue")
		return
	}
	DeleteQueue(id)
}

func TestReadTimeout(t *testing.T) {
	id, _ := CreateQueue(queueName)
	Enqueue(id, object)
	response, err := Read(id, 2)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
		return
	}
	if response.Id != id || !bytes.Equal(response.Object, object) {
		t.Errorf("want %q | %q, got %q | %q", response.Id, response.Object, id, object)
		return
	}
	// Simulate processing failure
	time.Sleep(time.Duration(3) * time.Second)

	// Object should now be ready to read again
	response, err = Read(id, 2)
	if err != nil {
		t.Errorf("unexpected read error: %v", err)
		return
	}
	if response.Id != id || !bytes.Equal(response.Object, object) {
		t.Errorf("want %q | %q, got %q | %q", response.Id, response.Object, id, object)
		return
	}
	if err := Dequeue(id, response.EntityId); err != nil {
		t.Errorf("unexpected dequeue error: %v", err)
		return
	}

	_, err = Read(id, 2)
	if err == nil {
		t.Errorf("expected read error on empty queue")
		return
	}
	DeleteQueue(id)
}

func TestReadWithoutEnqueue(t *testing.T) {
	id, _ := CreateQueue(queueName)
	_, err := Read(id, 2)
	if err == nil {
		t.Errorf("expected read error on empty queue")
		return
	}
}

func TestDequeueWithoutRead(t *testing.T) {
	id, _ := CreateQueue(queueName)
	Enqueue(id, object)
	if err := Dequeue(id, "0"); err == nil {
		t.Errorf("expected dequeue error without read")
		return
	}
}

// TODO: concurrent enqueue
// TODO: concurrent dequeue
// TODO: concurrent enqueue/dequeue
