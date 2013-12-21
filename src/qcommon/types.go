package qcommon

type QueueId	string
type Object	[]byte

// Common type definitions to simplify json marshal/unmarshaling
// between client and server.
type IdData struct {
	Id	QueueId
}

type IdObjectData struct {
	Id	QueueId
	Object	[]byte
}
