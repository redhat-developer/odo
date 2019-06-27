package testing

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	volumeclient "github.com/libopenstorage/openstorage/api/client/volume"
	"github.com/libopenstorage/openstorage/api/server"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/stretchr/testify/assert"
	"go.pedge.io/dlog"

	"github.com/golang/mock/gomock"
)

const (
	host       = "http://127.0.0.1"
	mgmtPort   = 2376
	pluginPort = 2377
	driver     = "mock"
	version    = "v1"
)

// Init function to setup the http server

func init() {
	startServer()
}

// contructs the base url with host and port
func getBaseURL() string {
	return host + ":" + strconv.Itoa(mgmtPort)
}

func startServer() {

	if err := server.StartVolumeMgmtAPI(
		driver,
		volume.DriverAPIBase,
		mgmtPort,
	); err != nil {
		dlog.Errorf("Error starting the server")
	}

	// adding sleep to avoid race condition of connection refused.
	time.Sleep(1 * time.Second)
}

func TestVolumeCreateSuccess(t *testing.T) {

	baseURL := getBaseURL()
	var err error
	ts := newTestServer(driver)
	defer ts.Stop()

	// create a request
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)
	assert.NotNil(t, ts.client)

	// Setup request
	name := "myvol"
	size := uint64(1234)

	req := &api.VolumeCreateRequest{
		Locator: &api.VolumeLocator{Name: name},
		Source:  &api.Source{},
		Spec:    &api.VolumeSpec{Size: size},
	}

	// Setup mock functions
	id := "myid"
	ts.MockDriver().
		EXPECT().
		Create(req.GetLocator(), req.GetSource(), req.GetSpec()).
		Return(id, nil)

	// create a volume client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.Nil(t, err)
	assert.Equal(t, id, res)
}

func TestVolumeCreateFailed(t *testing.T) {
	var err error

	ts := newTestServer(driver)

	defer ts.Stop()

	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)
	assert.NotNil(t, ts.client)

	req := &api.VolumeCreateRequest{}

	// Setup mock functions
	ts.MockDriver().
		EXPECT().
		Create(req.GetLocator(), req.GetSource(), req.GetSpec()).
		Return("", fmt.Errorf("error in create"))

	// create a volume client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Create(req.GetLocator(), req.GetSource(), req.GetSpec())

	assert.NotNil(t, err)
	assert.EqualValues(t, "", res)
	assert.Contains(t, err.Error(), "error in create")
}

func TestVolumeDeleteSuccess(t *testing.T) {
	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	// Setup mock
	id := "myid"

	ts.MockDriver().
		EXPECT().
		Delete(id).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	err = driverclient.Delete(id)
	assert.Nil(t, err)
}

func TestVolumeDeleteFailed(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	// Setup mock

	id := "myid"

	ts.MockDriver().
		EXPECT().
		Delete(id).
		Return(fmt.Errorf("error in delete"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	err = driverclient.Delete(id)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in delete")
}

func TestVolumeSnapshotCreateSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	id := "myid"
	name := "snapName"

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	req := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: name},
		Readonly: true,
	}

	//mock Snapshot call
	ts.MockDriver().
		EXPECT().
		Snapshot(req.GetId(), req.GetReadonly(), req.GetLocator()).
		Return(id, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Snapshot(req.GetId(), req.GetReadonly(), req.GetLocator())

	assert.Nil(t, err)
	assert.EqualValues(t, id, res)

}

func TestVolumeSnapshotCreateFailed(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	id := "myid"
	name := "snapName"

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	req := &api.SnapCreateRequest{Id: id,
		Locator:  &api.VolumeLocator{Name: name},
		Readonly: true,
	}

	//mock Snapshot call
	ts.MockDriver().
		EXPECT().
		Snapshot(req.GetId(), req.GetReadonly(), req.GetLocator()).
		Return("", fmt.Errorf("error in snapshot create"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Snapshot(req.GetId(), req.GetReadonly(), req.GetLocator())

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in snapshot create")
	assert.EqualValues(t, "", res)

}

func TestVolumeInspectSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	id := "myid"
	var size uint64 = 1234
	name := "inspectVol"

	ts.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{
			&api.Volume{
				Id: id,
				Locator: &api.VolumeLocator{
					Name: name,
				},
				Spec: &api.VolumeSpec{
					Size:      size,
					Encrypted: true,
					Shared:    false,
					Format:    api.FSType_FS_TYPE_EXT4,
					HaLevel:   3,
				},
			},
		}, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Inspect([]string{id})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, res)
	assert.EqualValues(t, res[0].GetId(), id)
	assert.EqualValues(t, false, res[0].GetSpec().GetShared())
	assert.EqualValues(t, 3, res[0].GetSpec().GetHaLevel())

}

func TestVolumeInspectFailed(t *testing.T) {
	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	id := "myid"

	ts.MockDriver().
		EXPECT().
		Inspect([]string{id}).
		Return([]*api.Volume{},
			fmt.Errorf("error in inspect"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Inspect([]string{id})

	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestVolumeSetSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	// create a volume request

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Set(id, req.GetLocator(), req.GetSpec()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create driver client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res := driverclient.Set(id, req.GetLocator(), req.GetSpec())
	assert.Nil(t, res)

}

func TestVolumeSetFailed(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	// create a volume request

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:  api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	ts.MockDriver().
		EXPECT().
		Set(id, req.GetLocator(), req.GetSpec()).
		Return(fmt.Errorf("error in set"))

	// create driver client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res := driverclient.Set(id, req.GetLocator(), req.GetSpec())

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in set")
}

func TestVolumeAttachSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Attach(id, gomock.Any()).
			Return("", nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create driver client
	driverclient := volumeclient.VolumeDriver(ts.client)

	_, err = driverclient.Attach(id, req.GetOptions())

	assert.Nil(t, err)

}

func TestVolumeAttachFailed(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	ts.MockDriver().
		EXPECT().
		Attach(id, gomock.Any()).
		Return("", fmt.Errorf("some error"))

	// create driver client
	driverclient := volumeclient.VolumeDriver(ts.client)

	_, err = driverclient.Attach(id, req.GetOptions())

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "some error")

}

func TestVolumeDetachSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(
		ts.MockDriver().
			EXPECT().
			Detach(id, gomock.Any()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Detach(id, req.GetOptions())

	assert.Nil(t, res)
}

func TestVolumeDetachFailed(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")

	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach: api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF,
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	ts.MockDriver().
		EXPECT().
		Detach(id, gomock.Any()).
		Return(fmt.Errorf("Error in detaching"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Detach(id, req.GetOptions())

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "Error in detaching")
}

func TestVolumeMountSuccess(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach:    api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			MountPath: "/mnt",
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(

		ts.MockDriver().
			EXPECT().
			Mount(id, gomock.Any(), gomock.Any()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	//create driverclient
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Mount(id, req.GetAction().GetMountPath(), req.GetOptions())
	assert.Nil(t, res)
}

func TestVolumeMountFailedNoMountPath(t *testing.T) {

	ts := newTestServer(driver)

	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(10)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Attach:    api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_ON,
			MountPath: "",
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	//create driverclient
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Mount(id, req.GetAction().GetMountPath(), req.GetOptions())
	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "Invalid mount path")
}

func TestVolumeStatsSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	bytesUsed := uint64(1234)
	writeBytes := uint64(1234)

	id := "myid"
	//req := &api.Stats{BytesUsed: bytesUsed}

	ts.MockDriver().
		EXPECT().
		Stats(id, gomock.Any()).
		Return(
			&api.Stats{
				BytesUsed:  bytesUsed,
				WriteBytes: writeBytes,
			},
			nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Stats(id, true)

	assert.Nil(t, err)
	assert.Equal(t, bytesUsed, res.BytesUsed)

}

func TestVolumeStatsFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	id := "myid"

	stats := &api.Stats{}

	ts.MockDriver().
		EXPECT().
		Stats(id, true).
		Return(stats,
			fmt.Errorf("Failed to get stats"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)

	res, err := driverclient.Stats(id, true)

	assert.NotNil(t, err)
	assert.ObjectsAreEqualValues(stats, res)
	//assert.Contains(t, err.Error(), "Failed to get stats")
}

func TestVolumeUnmountSuccess(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(1000)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF,
			MountPath: "/mnt",
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	gomock.InOrder(

		ts.MockDriver().
			EXPECT().
			Unmount(id, gomock.Any(), gomock.Any()).
			Return(nil),

		ts.MockDriver().
			EXPECT().
			Inspect([]string{id}).
			Return([]*api.Volume{
				&api.Volume{
					Id: id,
					Locator: &api.VolumeLocator{
						Name: name,
					},
					Spec: &api.VolumeSpec{
						Size: size,
					},
				},
			}, nil),
	)

	// setup client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Unmount(id, req.GetAction().GetMountPath(), req.GetOptions())

	assert.Nil(t, res)

}

func TestVolumeUnmountFailed(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	name := "myvol"
	id := "myid"
	size := uint64(1000)

	//create request
	req := &api.VolumeSetRequest{
		Options: map[string]string{},
		Action: &api.VolumeStateAction{
			Mount:     api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF,
			MountPath: "/mnt",
		},
		Locator: &api.VolumeLocator{Name: name},
		Spec:    &api.VolumeSpec{Size: size},
	}

	ts.MockDriver().
		EXPECT().
		Unmount(id, gomock.Any(), gomock.Any()).
		Return(fmt.Errorf("error in unmount"))

	// setup client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Unmount(id, req.GetAction().GetMountPath(), req.GetOptions())

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in unmount")
}

func TestVolumeQuiesceSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// volume instance
	id := "myid"
	//	name := "name"
	//	size := uint64(1234)
	quiesceid := "qid"
	timeout := uint64(5)

	ts.MockDriver().
		EXPECT().
		Quiesce(id, timeout, quiesceid).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Quiesce(id, timeout, quiesceid)

	assert.Nil(t, res)

}

func TestVolumeQuiesceFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// volume instance
	id := "myid"
	quiesceid := "qid"
	timeout := uint64(5)

	ts.MockDriver().
		EXPECT().
		Quiesce(id, timeout, quiesceid).
		Return(fmt.Errorf("error in quiesce"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Quiesce(id, timeout, quiesceid)

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in quiesce")

}

func TestVolumeUnquiesceSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	id := "myid"

	ts.MockDriver().
		EXPECT().
		Unquiesce(id).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Unquiesce(id)

	assert.Nil(t, res)
}

func TestVolumeUnquiesceFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	id := "myid"

	ts.MockDriver().
		EXPECT().
		Unquiesce(id).
		Return(fmt.Errorf("error in unquiesce"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Unquiesce(id)

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in unquiesce")
}

func TestVolumeRestoreSuccess(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	snapID := "snapid"
	volID := "volid"

	ts.MockDriver().
		EXPECT().
		Restore(volID, snapID).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Restore(volID, snapID)

	assert.Nil(t, res)
}

func TestVolumeRestoreFailed(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	snapID := "snapid"
	volID := "volid"

	ts.MockDriver().
		EXPECT().
		Restore(volID, snapID).
		Return(fmt.Errorf("error in restore"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res := driverclient.Restore(volID, snapID)

	assert.NotNil(t, res)
	assert.Contains(t, res.Error(), "error in restore")
}

func TestVolumeUsedSizeSuccess(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	volID := "myid"
	usedSize := uint64(1234)

	ts.MockDriver().
		EXPECT().
		UsedSize(volID).
		Return(usedSize, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.UsedSize(volID)

	assert.Nil(t, err)
	assert.Equal(t, usedSize, res)

}

func TestVolumeUsedSizeFailed(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	volID := "volid"
	usedSize := uint64(1234)

	ts.MockDriver().
		EXPECT().
		UsedSize(volID).
		Return(usedSize, fmt.Errorf("Failed to get used size"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.UsedSize(volID)

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), res)
	//	assert.Contains(t, err.Error(), "Failed to get used size")

}

func TestVolumeEnumerateSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create volume locator

	configLabel := make(map[string]string)
	configLabel["config1"] = "c1"

	name := "loc"
	vl := &api.VolumeLocator{
		Name: name,
		VolumeLabels: map[string]string{
			"dept": "auto",
			"sub":  "geo",
		},
	}

	id := "myid"
	size := uint64(1234)

	ts.MockDriver().
		EXPECT().
		Enumerate(vl, configLabel).
		Return([]*api.Volume{
			&api.Volume{
				Id:      id,
				Locator: vl,
				Spec: &api.VolumeSpec{
					Size: size,
				},
			},
		}, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.Enumerate(vl, configLabel)

	assert.Nil(t, err)
	assert.NotNil(t, res)

	if res != nil {
		assert.EqualValues(t, id, res[0].GetId())
	}
}

func TestVolumeEnumerateFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()
	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create volume locator

	configLabel := make(map[string]string)
	configLabel["config1"] = "cnfig1"

	name := "vol"
	vl := &api.VolumeLocator{
		Name: name,
		VolumeLabels: map[string]string{
			"class": "f9",
		},
	}

	ts.MockDriver().
		EXPECT().
		Enumerate(vl, configLabel).
		Return([]*api.Volume{},
			fmt.Errorf("error in enumerate"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.Enumerate(vl, configLabel)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in enumerate")
	assert.Empty(t, res)

}

func TestVolumeSnapshotEnumerateSuccess(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	ids := []string{
		"snapid1",
		"snapid2",
	}

	snapLabels := map[string]string{
		"dept": "auto",
		"sub":  "geo",
	}

	ts.MockDriver().
		EXPECT().
		SnapEnumerate(ids, snapLabels).
		Return([]*api.Volume{
			&api.Volume{
				Id: ids[0],
				Locator: &api.VolumeLocator{
					Name: "snap1",
				},
			},
			&api.Volume{
				Id: ids[1],
				Locator: &api.VolumeLocator{
					Name: "snap2",
				},
			},
		}, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.SnapEnumerate(ids, snapLabels)

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 2)

}

func TestVolumeSnapshotEnumerateFailed(t *testing.T) {
	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	ids := []string{
		"snapid1",
		"snapid2",
	}

	snapLabels := map[string]string{
		"dept": "auto",
		"sub":  "geo",
	}

	ts.MockDriver().
		EXPECT().
		SnapEnumerate(ids, snapLabels).
		Return([]*api.Volume{},
			fmt.Errorf("error in snap enumerate"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.SnapEnumerate(ids, snapLabels)

	assert.NotNil(t, err)
	assert.Empty(t, res)

}
func TestVolumeGetActiveRequestsSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	acreqs := &api.ActiveRequests{
		ActiveRequest: []*api.ActiveRequest{
			&api.ActiveRequest{
				ReqestKV: map[int64]string{
					1: "vol1",
				},
			},
			&api.ActiveRequest{
				ReqestKV: map[int64]string{
					2: "vol2",
				},
			},
		},
		RequestCount: 2,
	}

	ts.MockDriver().
		EXPECT().
		GetActiveRequests().
		Return(acreqs, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.GetActiveRequests()

	assert.Nil(t, err)
	assert.EqualValues(t, 2, res.GetRequestCount())
}

func TestVolumeGetActiveRequestsFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	ts.MockDriver().
		EXPECT().
		GetActiveRequests().
		Return(nil, fmt.Errorf("error in active requests"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.GetActiveRequests()

	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error in active requests")
}

func TestCredsCreateSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]string{
		"c1": "cred1",
		"c2": "cred2",
	}

	// Creata cred request
	cred := &api.CredCreateRequest{
		InputParams: credsmap,
	}

	ts.MockDriver().
		EXPECT().
		CredsCreate(cred.InputParams).
		Return("dummy-uuid", nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.CredsCreate(credsmap)

	assert.Nil(t, err)
	assert.EqualValues(t, "dummy-uuid", res)
}

func TestCredsCreateFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]string{
		"c1": "cred1",
		"c2": "cred2",
	}

	// Creata cred request
	cred := &api.CredCreateRequest{
		InputParams: credsmap,
	}

	ts.MockDriver().
		EXPECT().
		CredsCreate(cred.InputParams).
		Return("", fmt.Errorf("error in creds create"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.CredsCreate(credsmap)

	assert.NotNil(t, err)
	assert.EqualValues(t, "", res)
	assert.Contains(t, err.Error(), "error in creds create")
}

func TestCredsEnumerateSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]interface{}{
		"c1": "cred1",
		"c2": "cred2",
	}

	ts.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(credsmap, nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.CredsEnumerate()

	assert.Nil(t, err)
	assert.NotEmpty(t, res)
	assert.EqualValues(t, "cred1", res["c1"])
}

func TestCredsEnumerateFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// create a Creds request
	credsmap := map[string]interface{}{}

	ts.MockDriver().
		EXPECT().
		CredsEnumerate().
		Return(credsmap, fmt.Errorf("error in creds enumerate"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	res, err := driverclient.CredsEnumerate()

	assert.NotNil(t, err)
	assert.Empty(t, res)
	//assert.Contains(t, err.Error(), "error in creds enumerate")
}

func TestCredsValidateSuccess(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// cred uuid
	uuid := "dummy-validate-1101-uuid"

	ts.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(nil)

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	err = driverclient.CredsValidate(uuid)

	assert.Nil(t, err)
}

func TestCredsValidateFailed(t *testing.T) {

	ts := newTestServer(driver)
	defer ts.Stop()

	var err error
	baseURL := getBaseURL()

	ts.client, err = volumeclient.NewDriverClient(baseURL, driver, version, "")
	assert.Nil(t, err)

	// cred uuid
	uuid := "dummy-validate-1101-uuid"

	ts.MockDriver().
		EXPECT().
		CredsValidate(uuid).
		Return(fmt.Errorf("error in creds validate"))

	// create client
	driverclient := volumeclient.VolumeDriver(ts.client)
	err = driverclient.CredsValidate(uuid)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error in creds validate")
}
