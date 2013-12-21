package qclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

func doDequeue(readResponse *ReadResponse) error{
	resp, err := http.PostForm(Host + "/dequeue", url.Values{"id": {string(readResponse.Id)}, "entityId": {string(readResponse.EntityId)}})
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

func Read(id QueueId, timeout int) (*ReadResponse, error) {
	resp, err := http.PostForm(Host + "/read", url.Values{"id": {string(id)}})
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

	readResponse := new(ReadResponse)
	err = json.Unmarshal(body, &readResponse)
	if err != nil {
		return nil, err
	}

	key := ActiveReadKey{id: readResponse.Id, entityId: readResponse.EntityId,}
	if _, present := activeReads[key]; present {
		return nil, fmt.Errorf("Attempt to read already active read")
	}

	if err = doDequeue(readResponse); err != nil {
		return readResponse, err
	}

	t := time.AfterFunc(time.Duration(timeout) * time.Second, func() { readTimeout(readResponse) })
	activeReads[key] = t

	return readResponse, nil
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

