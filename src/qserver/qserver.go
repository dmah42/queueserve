package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	// TODO: flag
	port = 4242
)

var (
	queues = map[string](*queue){}
)

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.ParseForm() != nil {
		http.Error(w, "Unable to parse form values", http.StatusBadRequest)
		return
	}

	name := r.Form["name"][0]
	if _, present := queues[name]; present {
		http.Error(w, "Queue already exists", http.StatusConflict)
		return
	}

	log.Printf("creating queue %q", name)
	queues[name] = newQueue()

	idData := struct { Id string }{Id: name}
	b, err := json.Marshal(idData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.ParseForm() != nil {
		http.Error(w, "Unable to parse form values", http.StatusBadRequest)
		return
	}

	name := r.Form["name"][0]
	if _, present := queues[name]; !present {
		http.Error(w, "Queue doesn't exist", http.StatusNotFound)
		return
	}

	log.Printf("getting queue %q", name)
	idData := struct { Id string }{Id: name}
	b, err := json.Marshal(idData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.ParseForm() != nil {
		http.Error(w, "Unable to parse form values", http.StatusBadRequest)
		return
	}

	id := r.Form["id"][0]
	if _, present := queues[id]; !present {
		http.Error(w, "Queue doesn't exist", http.StatusNotFound)
		return
	}

	log.Printf("deleting queue %q", id)
	delete(queues, id)
	w.WriteHeader(http.StatusOK)
}

func enqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.ParseForm() != nil {
		http.Error(w, "Unable to parse form values", http.StatusBadRequest)
		return
	}

	id := r.Form["id"][0]
	q, present := queues[id];
	if !present {
		http.Error(w, fmt.Sprintf("Queue %q doesn't exist", id), http.StatusNotFound)
		return
	}

	object := r.Form["object"][0]
	log.Printf("enqueue %q %q", id, object)
	q.enqueue([]byte(object))
	w.WriteHeader(http.StatusOK)
}

func dequeueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.ParseForm() != nil {
		http.Error(w, "Unable to parse form values", http.StatusBadRequest)
		return
	}

	id := r.Form["id"][0]
	q, present := queues[id];
	if !present {
		http.Error(w, fmt.Sprintf("Queue %q doesn't exist", id), http.StatusNotFound)
		return
	}
	object, valid := q.dequeue()
	if !valid {
		http.Error(w, "Attempt to dequeue from empty queue", http.StatusNotFound)
		return
	}

	log.Printf("dequeue %q %q", id, object)

	// TODO: shared objects
	idObjectData := struct {
		Id	string
		Object	[]byte
	} {
		Id:	id,
		Object:	object,
	}
	b, err := json.Marshal(idObjectData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func main() {
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/enqueue", enqueueHandler)
	http.HandleFunc("/dequeue", dequeueHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
