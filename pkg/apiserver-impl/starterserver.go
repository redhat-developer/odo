package apiserver_impl

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"

	"k8s.io/klog"

	openapi "github.com/redhat-developer/odo/pkg/apiserver-gen/go"
	"github.com/redhat-developer/odo/pkg/apiserver-impl/sse"
	"github.com/redhat-developer/odo/pkg/informer"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/feature"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

//go:embed ui/*
var staticFiles embed.FS

// swagger UI files are from https://github.com/swagger-api/swagger-ui/tree/master/dist

//go:embed swagger-ui/*
var swaggerFiles embed.FS

type ApiServer struct {
	PushWatcher <-chan struct{}
}

func StartServer(
	ctx context.Context,
	cancelFunc context.CancelFunc,
	randomPort bool,
	port int,
	devfilePath string,
	devfileFiles []string,
	fsys filesystem.Filesystem,
	kubernetesClient kclient.ClientInterface,
	podmanClient podman.Client,
	stateClient state.Client,
	preferenceClient preference.Client,
	informerClient *informer.InformerClient,
) (ApiServer, error) {
	pushWatcher := make(chan struct{})
	defaultApiService := NewDefaultApiService(
		cancelFunc,
		pushWatcher,
		kubernetesClient,
		podmanClient,
		stateClient,
		preferenceClient,
	)
	defaultApiController := openapi.NewDefaultApiController(defaultApiService)
	devstateApiService := NewDevstateApiService(
		cancelFunc,
		pushWatcher,
		kubernetesClient,
		podmanClient,
		stateClient,
		preferenceClient,
	)
	devstateApiController := openapi.NewDevstateApiController(devstateApiService)

	sseNotifier, err := sse.NewNotifier(ctx, fsys, devfilePath, devfileFiles)
	if err != nil {
		return ApiServer{}, err
	}

	router := openapi.NewRouter(sseNotifier, defaultApiController, devstateApiController)

	fSysSwagger, err := fs.Sub(swaggerFiles, "swagger-ui")
	if err != nil {
		// Assertion, error can only happen if the path "swagger-ui" is not valid
		panic(err)
	}
	swaggerServer := http.FileServer(http.FS(fSysSwagger))
	router.PathPrefix("/swagger-ui/").Handler(http.StripPrefix("/swagger-ui/", swaggerServer))

	if feature.IsEnabled(ctx, feature.UIServer) {
		var fSys fs.FS
		fSys, err = fs.Sub(staticFiles, "ui")
		if err != nil {
			// Assertion, error can only happen if the path "ui" is not valid
			panic(err)
		}

		staticServer := http.FileServer(http.FS(fSys))
		router.PathPrefix("/").Handler(staticServer)
	}

	if port == 0 && !randomPort {
		port, err = util.NextFreePort(20000, 30001, nil, "")
		if err != nil {
			klog.V(0).Infof("Unable to start the API server; encountered error: %v", err)
			cancelFunc()
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return ApiServer{}, fmt.Errorf("unable to start API Server listener on port %d: %w", port, err)
	}

	server := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Handler: router,
	}
	var errChan = make(chan error)
	go func() {
		errChan <- server.Serve(listener)
	}()

	// Get the actual port value assigned by the operating system
	listeningPort := listener.Addr().(*net.TCPAddr).Port
	if port != 0 && port != listeningPort {
		panic(fmt.Sprintf("requested port (%d) not the same as the actual port the API Server is bound to (%d)", port, listeningPort))
	}

	err = stateClient.SetAPIServerPort(ctx, listeningPort)
	if err != nil {
		klog.V(0).Infof("Unable to start the API server; encountered error: %v", err)
		cancelFunc()
	}

	if feature.IsEnabled(ctx, feature.UIServer) {
		info := fmt.Sprintf("Web console accessible at http://localhost:%d/", listeningPort)
		log.Spinner(info).End(true)
		informerClient.AppendInfo(info + "\n")
	}
	log.Spinner(fmt.Sprintf("API Server started at http://localhost:%d/api/v1", listeningPort)).End(true)
	log.Spinner(fmt.Sprintf("API documentation accessible at http://localhost:%d/swagger-ui/", listeningPort)).End(true)

	go func() {
		select {
		case <-ctx.Done():
			klog.V(0).Infof("Shutting down the API server: %v", ctx.Err())
			err = server.Shutdown(ctx)
			if err != nil {
				klog.V(1).Infof("Error while shutting down the API server: %v", err)
			}
		case err = <-errChan:
			klog.V(0).Infof("Stopping the API server; encountered error: %v", err)
			cancelFunc()
		}
	}()

	return ApiServer{
		PushWatcher: pushWatcher,
	}, nil
}
