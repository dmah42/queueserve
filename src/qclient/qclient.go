package qclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

type QueueId string
type QueueEntityId string
type Object []byte

type ReadResponse struct {
	Id	QueueId
	EntityId	QueueEntityId
	Object	Object
}

type ActiveReadKey struct {
	id	QueueId
	entityId	QueueEntityId
}

const (
	// TODO: flag
	port = 4242
	nullId = ""
)

var (
	Host = fmt.Sprintf("http://localhost:%d", port)
	activeReads = map[ActiveReadKey](*time.Timer){}
	queueEntityId int32 = 0
)

func CreateQueue(name string) (QueueId, error) {
	resp, err := http.PostForm(Host + "/create", url.Values{"name": {name}})
	if err != nil {
		return nullId, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nullId, fmt.Errorf(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nullId, err
	}

	idData := new(struct { Id string })
	err = json.Unmarshal(body, &idData)
	if err != nil {
		return nullId, err
	}
	return QueueId(idData.Id), nil
}

func GetQueue(name string) (QueueId, error) {
	resp, err := http.PostForm(Host + "/get", url.Values{"name": {name}})
	if err != nil {
		return nullId, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nullId, fmt.Errorf(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nullId, err
	}

	idData := new(struct { Id string })
	err = json.Unmarshal(body, &idData)
	if err != nil {
		return nullId, err
	}
	return QueueId(idData.Id), nil
}

func DeleteQueue(id QueueId) error {
	resp, err := http.PostForm(Host + "/delete", url.Values{"id": {string(id)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func Enqueue(id QueueId, object Object) error {
	resp, err := http.PostForm(Host + "/enqueue", url.Values{"id": {string(id)}, "object": {string(object)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func readTimeout(readResponse *ReadResponse) {
	Enqueue(readResponse.Id, readResponse.Object)
	key := ActiveReadKey{id: readResponse.Id, entityId: readResponse.EntityId,}
	delete(activeReads, key)
}

// Read actually dequeues from the server and then sets a timeout to re-enqueue the same object.
// Dequeue can then be implemented as a cancelation of the timeout. This ensures that the same
// object won't be dequeued from the server while it is being read.
func Read(id QueueId, timeout int) (*ReadResponse, error) {
	resp, err := http.PostForm(Host + "/dequeue", url.Values{"id": {string(id)}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	idObjectData := new(struct {
		Id	QueueId
		Object	[]byte
	})
	err = json.Unmarshal(body, &idObjectData)
	if err != nil {
		return nil, err
	}

	entityId := QueueEntityId(fmt.Sprintf("%d", atomic.AddInt32(&queueEntityId, 1)))
	readResponse := ReadResponse{
		Id:		idObjectData.Id,
		EntityId:	entityId,
		Object:		Object(idObjectData.Object),
	}
	key := ActiveReadKey{id: readResponse.Id, entityId: readResponse.EntityId}
	if _, present := activeReads[key]; present {
		return nil, fmt.Errorf("Attempt to read already active read")
	}

	t := time.AfterFunc(time.Duration(timeout) * time.Second, func() { readTimeout(&readResponse) })
	activeReads[key] = t

	return &readResponse, nil
}

func Dequeue(id QueueId, entityId QueueEntityId) error {
	key := ActiveReadKey{id: id, entityId: entityId,}
	timer, present := activeReads[key]
	if !present {
		return fmt.Errorf("Attempt to dequeue item without reading")
	}
	timer.Stop()
	delete(activeReads, key)
	return nil
}

