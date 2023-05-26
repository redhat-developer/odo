package apiserver_impl

import (
	"context"
	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"k8s.io/klog"
	"log"
	"net/http"
)

func StartServer(ctx context.Context, cancelFunc context.CancelFunc) {
	log.Printf("Server started")

	DefaultApiService := NewDefaultApiService()
	DefaultApiController := openapi.NewDefaultApiController(DefaultApiService)

	router := openapi.NewRouter(DefaultApiController)
	server := &http.Server{Addr: ":20000", Handler: router}
	var errChan = make(chan error)
	go func() {
		err := server.ListenAndServe()
		errChan <- err
	}()
	go func() {
		select {
		case <-ctx.Done():
			err := server.Shutdown(ctx)
			klog.V(1).Infof("Shutting down the server: %v", ctx.Err())
			if err != nil {
				klog.V(1).Infof("Error while shutting down the server: %v", err)
			}
		case err := <-errChan:
			log.Printf("Stopping the server; encounter error: %v", err)
			cancelFunc()
		}
	}()
}
