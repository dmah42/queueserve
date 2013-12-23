package qclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"qcommon"
	"sync/atomic"
	"time"
)

type QueueEntityId string

type ReadResponse struct {
	Id	qcommon.QueueId
	EntityId	QueueEntityId
	Object	qcommon.Object
}

type ActiveReadKey struct {
	id	qcommon.QueueId
	entityId	QueueEntityId
}

const (
	nullId = ""
)

var (
	Port = 4242
	Host = "localhost"
	activeReads = map[ActiveReadKey](*time.Timer){}
	queueEntityId int32 = 0
)

func apiUrl(path string) string {
	return fmt.Sprintf("http://%s:%d/%s", Host, Port, path)
}

func getBody(path string, values url.Values) ([]byte, error) {
	resp, err := http.PostForm(apiUrl(path), values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}

	return body, nil
}

func CreateQueue(name string) (qcommon.QueueId, error) {
	body, err := getBody("create", url.Values{"name": {name}})
	if err != nil {
		return nullId, err
	}

	idData := new(qcommon.IdData)
	err = json.Unmarshal(body, &idData)
	if err != nil {
		return nullId, err
	}
	return idData.Id, nil
}

func GetQueue(name string) (qcommon.QueueId, error) {
	body, err := getBody("get", url.Values{"name": {name}})
	if err != nil {
		return nullId, err
	}

	idData := new(qcommon.IdData)
	err = json.Unmarshal(body, &idData)
	if err != nil {
		return nullId, err
	}
	return idData.Id, nil
}

func DeleteQueue(id qcommon.QueueId) error {
	_, err := getBody("delete", url.Values{"id": {string(id)}})
	return err
}

func Enqueue(id qcommon.QueueId, object qcommon.Object) error {
	_, err := getBody("enqueue", url.Values{"id": {string(id)}, "object": {string(object)}})
	return err
}

func readTimeout(readResponse *ReadResponse) {
	Enqueue(readResponse.Id, readResponse.Object)
	key := ActiveReadKey{id: readResponse.Id, entityId: readResponse.EntityId,}
	delete(activeReads, key)
}

// Read actually dequeues from the server and then sets a timeout to re-enqueue the same object.
// Dequeue can then be implemented as a cancelation of the timeout. This ensures that the same
// object won't be dequeued from the server while it is being read.
func Read(id qcommon.QueueId, timeout time.Duration) (*ReadResponse, error) {
	body, err := getBody("dequeue", url.Values{"id": {string(id)}})
	if err != nil {
		return nil, err
	}

	idObjectData := new(qcommon.IdObjectData)
	err = json.Unmarshal(body, &idObjectData)
	if err != nil {
		return nil, err
	}

	if idObjectData.Id != id {
		return nil, fmt.Errorf("Mismatch queue ids: %q vs %q", idObjectData.Id, id)
	}

	entityId := QueueEntityId(fmt.Sprintf("%d", atomic.AddInt32(&queueEntityId, 1)))
	readResponse := ReadResponse{
		Id:		idObjectData.Id,
		EntityId:	entityId,
		Object:		idObjectData.Object,
	}
	key := ActiveReadKey{id: readResponse.Id, entityId: readResponse.EntityId}
	if _, present := activeReads[key]; present {
		return nil, fmt.Errorf("Attempt to read already active read")
	}

	t := time.AfterFunc(timeout, func() { readTimeout(&readResponse) })
	activeReads[key] = t

	return &readResponse, nil
}

func Dequeue(id qcommon.QueueId, entityId QueueEntityId) error {
	key := ActiveReadKey{id: id, entityId: entityId,}
	timer, present := activeReads[key]
	if !present {
		return fmt.Errorf("Attempt to dequeue item without reading")
	}
	if (!timer.Stop()) {
		return fmt.Errorf("Read timed out")
	}
	delete(activeReads, key)
	return nil
}

