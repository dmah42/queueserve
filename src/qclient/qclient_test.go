package qclient

import (
	"bytes"
	"flag"
	"fmt"
	"sync"
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
	load = flag.Int("load", 1000, "number of concurrent events to test")
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

func TestDuplicateCreate(t *testing.T) {
	id, _ := CreateQueue(queueName)
	_, err := CreateQueue(queueName)
	if err == nil {
		t.Errorf("expected duplicate creation error")
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
	DeleteQueue(id)
}

func TestDequeueWithoutRead(t *testing.T) {
	id, _ := CreateQueue(queueName)
	Enqueue(id, object)
	if err := Dequeue(id, "0"); err == nil {
		t.Errorf("expected dequeue error without read")
		return
	}
	DeleteQueue(id)
}

// TODO: benchmark
func TestConcurrentEnqueue(t *testing.T) {
	id, _ := CreateQueue(queueName)

	errorChan := make(chan int)
	for i := 0; i < *load; i++ {
		s := fmt.Sprintf("%d", i)
		go func(i int) {
			err := Enqueue(id, []byte(s))
			if err != nil {
				errorChan <- 1
			}
			if i+1 == *load {
				errorChan <- -1
			}
		}(i)
	}

	var errorCount int
	done := false
	for {
		if done {
			break
		}

		select {
			case i := <-errorChan:
				if i < 0 {
					done = true
				} else {
					errorCount += i
				}
		}
	}

	if errorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
		return
	}
	DeleteQueue(id)
}

// TODO: benchmark
func TestConcurrentDequeue(t *testing.T) {
	id, _ := CreateQueue(queueName)

	var waitgroup sync.WaitGroup
	waitgroup.Add(*load)
	for i := 0; i < *load; i++ {
		s := fmt.Sprintf("%d", i)
		go func() {
			Enqueue(id, []byte(s))
			waitgroup.Done()
		}()
	}
	waitgroup.Wait()

	errorChan := make(chan int)
	for i := 0; i < *load; i++ {
		go func(i int) {
			resp, err := Read(id, 2)
			if err != nil {
				errorChan <- 1
			} else if err = Dequeue(id, resp.EntityId); err != nil {
				errorChan <- 1
			}
			if i+1 == *load {
				errorChan <- -1
			}
		}(i)
	}

	var errorCount int
	done := false
	for {
		if done {
			break
		}

		select {
			case i := <-errorChan:
				if i < 0 {
					done = true
				} else {
					errorCount += i
				}
		}
	}

	if errorCount != 0 {
		t.Errorf("Expected 0 errors, got %d", errorCount)
		return
	}
	DeleteQueue(id)
}

// TODO: concurrent enqueue/dequeue
