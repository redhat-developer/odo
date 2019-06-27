package server

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubernetes-csi/csi-test/utils"
	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/volume"
	vol_drivers "github.com/libopenstorage/openstorage/volume/drivers"
	mockdriver "github.com/libopenstorage/openstorage/volume/drivers/mock"
)

const (
	mockDriverName = "mock"
)

// testServer is a simple struct used abstract
// the creation and setup of the gRPC CSI service
type testServer struct {
	m  *mockdriver.MockVolumeDriver
	mc *gomock.Controller
}

func setupMockDriver(tester *testServer, t *testing.T) {
	vol_drivers.Add(mockDriverName, func(map[string]string) (volume.VolumeDriver, error) {
		return tester.m, nil
	})

	var err error

	// Register mock driver
	err = vol_drivers.Register(mockDriverName, nil)
	assert.Nil(t, err)
}

func newTestServer(t *testing.T) *testServer {
	tester := &testServer{}

	// Add driver to registry
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})
	tester.m = mockdriver.NewMockVolumeDriver(tester.mc)

	setupMockDriver(tester, t)
	return tester
}

func (s *testServer) MockDriver() *mockdriver.MockVolumeDriver {
	return s.m
}

func (s *testServer) Stop() {
	// Remove from registry
	vol_drivers.Remove(mockDriverName)

	// Check mocks
	s.mc.Finish()
}

func Setup(t *testing.T) (*httptest.Server, *testServer) {
	vapi := &volAPI{}
	router := mux.NewRouter()
	// Register all routes from the App
	for _, route := range vapi.Routes() {
		router.Methods(route.verb).
			Path(route.path).
			Name(mockDriverName).
			Handler(http.HandlerFunc(route.fn))
	}

	ts := httptest.NewServer(router)

	testVolDriver := newTestServer(t)
	return ts, testVolDriver
}

func TestClientCredCreate(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()
	var uuid string
	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)
	// S3 cloud provider
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "s3",
			api.OptCredRegion:    "east",
			api.OptCredEndpoint:  "s3.url.com",
			api.OptCredAccessKey: "s3accesskey",
			api.OptCredSecretKey: "s3secretekey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "s3",
			api.OptCredRegion:    "east",
			api.OptCredEndpoint:  "s3.url.com",
			api.OptCredAccessKey: "",
			api.OptCredSecretKey: "",
		}).
		Return("", fmt.Errorf("Missing s3 access/secrete keys")).
		Times(1)

	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "azure",
			api.OptCredAzureAccountName: "azuretest",
			api.OptCredAzureAccountKey:  "azureaccountkey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "azure",
			api.OptCredAzureAccountName: "",
			api.OptCredAzureAccountKey:  "",
		}).
		Return("", fmt.Errorf("Missing azure account name/keys")).
		Times(1)

	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "google",
			api.OptCredGoogleProjectID: "googletestproject",
			api.OptCredGoogleJsonKey:   "googlejsonkey",
		}).
		Return("gooduuid", nil).
		Times(1)
	testVolDriver.MockDriver().EXPECT().
		CredsCreate(map[string]string{api.OptCredType: "google",
			api.OptCredGoogleProjectID: "",
			api.OptCredGoogleJsonKey:   "",
		}).
		Return("", fmt.Errorf("Missing google project/json key")).
		Times(1)

		//Invoke CredsCreate for S3
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "s3",
		api.OptCredRegion:    "east",
		api.OptCredEndpoint:  "s3.url.com",
		api.OptCredAccessKey: "s3accesskey",
		api.OptCredSecretKey: "s3secretekey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "s3",
		api.OptCredRegion:    "east",
		api.OptCredEndpoint:  "s3.url.com",
		api.OptCredAccessKey: "",
		api.OptCredSecretKey: "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

	// Azure cloud provider
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "azure",
		api.OptCredAzureAccountName: "azuretest",
		api.OptCredAzureAccountKey:  "azureaccountkey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "azure",
		api.OptCredAzureAccountName: "",
		api.OptCredAzureAccountKey:  "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

	//Google

	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "google",
		api.OptCredGoogleProjectID: "googletestproject",
		api.OptCredGoogleJsonKey:   "googlejsonkey",
	})
	require.NoError(t, err)
	require.Equal(t, uuid, "gooduuid")
	uuid, err = client.VolumeDriver(cl).CredsCreate(map[string]string{api.OptCredType: "google",
		api.OptCredGoogleProjectID: "",
		api.OptCredGoogleJsonKey:   "",
	})
	require.Error(t, err)
	require.Equal(t, uuid, "")
	require.Contains(t, err.Error(), "Missing")

}

func TestClientCredsValidateAndDelete(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().CredsDelete("gooduuid").Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CredsDelete("baduuid").Return(fmt.Errorf("Invalid UUID")).Times(1)

	testVolDriver.MockDriver().EXPECT().CredsValidate("gooduuid").Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().CredsValidate("baduuid").Return(fmt.Errorf("Invalid UUID")).Times(1)

	// Delete creds
	err = client.VolumeDriver(cl).CredsDelete("gooduuid")
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CredsDelete("baduuid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid UUID")
	err = client.VolumeDriver(cl).CredsDelete("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")
	//Validate creds
	err = client.VolumeDriver(cl).CredsValidate("gooduuid")
	require.NoError(t, err)
	err = client.VolumeDriver(cl).CredsValidate("baduuid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid UUID")
	err = client.VolumeDriver(cl).CredsValidate("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "404")

}

func TestClientCredsList(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)
	enumerateData := make(map[string]interface{}, 0)
	testVolDriver.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(enumerateData, nil).
		Times(1)
	_, err = client.VolumeDriver(cl).CredsEnumerate()
	require.NoError(t, err)
}
