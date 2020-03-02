package testingutil

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"
)

// FakePortListener starts a fake test server and listens on the given localPort
func FakePortListener(startedChan chan<- bool, stopChan <-chan bool, localPort int) error {
	testServer := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("running")
	}))

	listener, err := net.Listen("tcp", "localhost:"+strconv.Itoa(localPort))
	if err != nil {
		return err
	}
	testServer.Listener = listener
	testServer.Start()
	startedChan <- true
	timeout := time.After(10 * time.Second)

	select {
	case <-stopChan:
		testServer.Close()
	case <-timeout:
		testServer.Close()
		return errors.New("timeout waiting for listerner to start")
	}
	return nil
}
