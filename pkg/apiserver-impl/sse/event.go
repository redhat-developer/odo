package sse

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type EventType int

const (
	Heartbeat EventType = iota + 1
	DevfileUpdated
)

type Event struct {
	eventType EventType
	data      interface{}
}

func (e Event) toSseString() (string, error) {
	var eventName string
	switch e.eventType {
	case Heartbeat:
		return ": heartbeat\n\n", nil
	case DevfileUpdated:
		eventName = "DevfileUpdated"
	default:
		return "", fmt.Errorf("unrecognized event type:%v", e.eventType)
	}

	if e.data == nil {
		return fmt.Sprintf("event: %s\n\n", eventName), nil
	}

	jsonBytes, err := json.Marshal(e.data)
	if err != nil {
		return "", err
	}
	var dataStr bytes.Buffer
	err = json.Compact(&dataStr, jsonBytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, dataStr.String()), nil
}
