package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"qclient"
	"time"
)

const readTimeout = time.Duration(1) * time.Second

var (
	port = flag.Int("port", 4242, "the port the server is listening on")
	host = flag.String("host", "localhost", "the host the server is running on")
	queue = flag.String("queue", "q", "the name of the queue that has been created")
	count = flag.Int("count", 100, "the number of operations to attempt")
)

func main() {
	flag.Parse()

	qclient.Host = *host
	qclient.Port = *port

	id, err := qclient.GetQueue(*queue)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < *count; i++ {
		if rand.Float32() > 0.5 {
			if err = qclient.Enqueue(id, []byte(fmt.Sprintf("object%d", i))); err != nil {
				log.Printf("enq: %v", err)
			}
		} else {
			resp, err := qclient.Read(id, readTimeout)
			if err != nil {
				log.Printf("read: %v", err)
				continue
			}
			if err = qclient.Dequeue(id, resp.EntityId); err != nil {
				log.Printf("deq: %v", err)
			}
		}
	}
}
