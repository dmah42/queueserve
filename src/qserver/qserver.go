package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"qcommon"
)

var (
	port = flag.Int("port", 4242, "port to listen on")
	verbose = flag.Bool("verbose", false, "verbose logging")
	queues = map[string](*queue){}
)

func vLog(format string, a ...interface{}) {
	if *verbose {
		log.Printf(format, a...)
	}
}

// returns the form value for the given key, or an error message/http status code pair.
func getFormValue(r *http.Request, key string) (string, int) {
	if r.Method != "POST" {
		return "POST only", http.StatusMethodNotAllowed
	}

	if r.ParseForm() != nil {
		return "Unable to parse form values", http.StatusBadRequest
	}

	if len(r.Form[key]) == 0 {
		return fmt.Sprintf("Missing key: %q", key), http.StatusBadRequest
	}

	return r.Form[key][0], http.StatusOK
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	name, status := getFormValue(r, "name")
	if status != http.StatusOK {
		http.Error(w, name, status)
		return
	}

	if _, present := queues[name]; present {
		http.Error(w, "Queue already exists", http.StatusConflict)
		return
	}

	vLog("creating queue %q", name)
	queues[name] = newQueue()

	idData := qcommon.IdData{Id: qcommon.QueueId(name)}
	b, err := json.Marshal(idData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	name, status := getFormValue(r, "name")
	if status != http.StatusOK {
		http.Error(w, name, status)
		return
	}

	if _, present := queues[name]; !present {
		http.Error(w, "Queue doesn't exist", http.StatusNotFound)
		return
	}

	vLog("getting queue %q", name)
	idData := qcommon.IdData{Id: qcommon.QueueId(name)}
	b, err := json.Marshal(idData)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, status := getFormValue(r, "id")
	if status != http.StatusOK {
		http.Error(w, id, status)
		return
	}

	if _, present := queues[id]; !present {
		http.Error(w, "Queue doesn't exist", http.StatusNotFound)
		return
	}

	vLog("deleting queue %q", id)
	delete(queues, id)
	w.WriteHeader(http.StatusOK)
}

func enqueueHandler(w http.ResponseWriter, r *http.Request) {
	id, status := getFormValue(r, "id")
	if status != http.StatusOK {
		http.Error(w, id, status)
		return
	}

	q, present := queues[id];
	if !present {
		http.Error(w, fmt.Sprintf("Queue %q doesn't exist", id), http.StatusNotFound)
		return
	}

	// TODO: extend getFormValue to take a slice of keys and return errors as appropriate.
	if len(r.Form["object"]) == 0 {
		http.Error(w, "Missing object field", http.StatusBadRequest)
		return
	}
	object := r.Form["object"][0]
	vLog("enqueue %q %q", id, object)
	q.enqueue([]byte(object))
	w.WriteHeader(http.StatusOK)
}

func dequeueHandler(w http.ResponseWriter, r *http.Request) {
	id, status := getFormValue(r, "id")
	if status != http.StatusOK {
		http.Error(w, id, status)
		return
	}

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

	vLog("dequeue %q %q", id, object)

	idObjectData := qcommon.IdObjectData{
		Id:	qcommon.QueueId(id),
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

	flag.Parse()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
