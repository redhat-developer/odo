package dev

import (
	"bytes"
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/watch"
	"log"
	"os"
	"testing"
)

func TestDev_Start(t *testing.T) {
	ctrl := gomock.NewController(t)
	d := DevClient{
		kubernetesClient: kclient.NewMockClientInterface(ctrl),
		watchClient:      watch.NewMockClient(ctrl),
	}
	//devfileObj, _ := devfile.ParseAndValidateFromFile("/home/dshah/src/odo/tests/examples/source/devfiles/nodejs/devfile.yaml")
	//devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	//devfileObj := parser.DevfileObj{
	//	Data: devfileData,
	//}
	out := bytes.Buffer{}
	path := "/home/dshah/src/nodejs-ex"
	platformContext := kubernetes.KubernetesContext{Namespace: "myproject"}
	_, _ = os.Getwd()
	os.Chdir("/home/dshah/src/nodejs-ex")
	devfileObj, _ := devfile.ParseAndValidateFromFile("./devfile.yaml")

	err := d.Start(devfileObj, platformContext, path, &out)
	log.Fatal(err)
}
