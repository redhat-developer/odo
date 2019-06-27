package server

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	client "github.com/libopenstorage/openstorage/api/client/volume"
)

func TestClientBackup(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().Backup(&api.BackupRequest{
		VolumeID:       "goodvol",
		CredentialUUID: "",
		Full:           false}).Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().Backup(&api.BackupRequest{
		VolumeID:       "badvol",
		CredentialUUID: "",
		Full:           false}).Return(fmt.Errorf("Volume not found")).Times(1)

	// Create Backup
	err = client.VolumeDriver(cl).
		Backup(&api.BackupRequest{
			VolumeID:       "goodvol",
			CredentialUUID: "",
			Full:           false})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		Backup(&api.BackupRequest{
			VolumeID:       "badvol",
			CredentialUUID: "",
			Full:           false})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Volume not found")
}

func TestClientBackupRestore(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupRestore(&api.BackupRestoreRequest{
		CloudBackupID:  "goodBackupid",
		CredentialUUID: ""}).
		Return(&api.BackupRestoreResponse{RestoreErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupRestore(&api.BackupRestoreRequest{
		CloudBackupID:  "badbackupid",
		CredentialUUID: ""}).
		Return(&api.BackupRestoreResponse{RestoreErr: "Backup not found"}).Times(1)

	//Invoke restore
	restoreResponse := client.VolumeDriver(cl).
		BackupRestore(&api.BackupRestoreRequest{
			CloudBackupID:  "goodBackupid",
			CredentialUUID: ""})
	require.Contains(t, restoreResponse.RestoreErr, "")
	restoreResponse = client.VolumeDriver(cl).
		BackupRestore(&api.BackupRestoreRequest{
			CloudBackupID:  "badbackupid",
			CredentialUUID: ""})
	require.Contains(t, restoreResponse.RestoreErr, "Backup not found")
}

func TestClientBackupDelete(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	goodInput := &api.BackupDeleteRequest{}
	goodInput.SrcVolumeID = "goodsrc"
	goodInput.CredentialUUID = ""

	badInput := &api.BackupDeleteRequest{}
	badInput.SrcVolumeID = "badsrc"
	badInput.CredentialUUID = ""
	testVolDriver.MockDriver().EXPECT().BackupDelete(goodInput).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupDelete(badInput).
		Return(fmt.Errorf("Src volume not found")).Times(1)
	//Invoke restore
	err = client.VolumeDriver(cl).
		BackupDelete(goodInput)
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		BackupDelete(badInput)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Src volume not found")
}

func TestClientBackupEnumerate(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)
	goodInput := &api.BackupEnumerateRequest{}
	goodInput.SrcVolumeID = ""
	goodInput.CredentialUUID = ""
	testVolDriver.MockDriver().EXPECT().BackupEnumerate(goodInput).
		Return(&api.BackupEnumerateResponse{EnumerateErr: ""}).Times(1)
	badInput := &api.BackupEnumerateRequest{}
	badInput.SrcVolumeID = ""
	badInput.CredentialUUID = ""
	testVolDriver.MockDriver().EXPECT().BackupEnumerate(badInput).
		Return(&api.BackupEnumerateResponse{EnumerateErr: "Credential invalid"}).Times(1)

	//Invoke Enumerate
	response := client.VolumeDriver(cl).
		BackupEnumerate(goodInput)
	require.Equal(t, response.EnumerateErr, "")
	response = client.VolumeDriver(cl).
		BackupEnumerate(badInput)
	require.Contains(t, response.EnumerateErr, "Credential invalid")
}

func TestClientBackupStatus(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupStatus(&api.BackupStsRequest{
		SrcVolumeID: "goodsrc"}).
		Return(&api.BackupStsResponse{StsErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupStatus(&api.BackupStsRequest{
		SrcVolumeID: "badsrc"}).
		Return(&api.BackupStsResponse{StsErr: "Invalid source volume"}).Times(1)

	//Invoke Enumerate
	response := client.VolumeDriver(cl).
		BackupStatus(&api.BackupStsRequest{
			SrcVolumeID: "goodsrc"})
	require.Equal(t, response.StsErr, "")
	response = client.VolumeDriver(cl).
		BackupStatus(&api.BackupStsRequest{
			SrcVolumeID: "badsrc"})
	require.Contains(t, response.StsErr, "Invalid source volume")
}

func TestClientBackupCatalogue(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupCatalogue(&api.BackupCatalogueRequest{
		CloudBackupID:  "goodcloudbackup",
		CredentialUUID: ""}).
		Return(&api.BackupCatalogueResponse{CatalogueErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupCatalogue(&api.BackupCatalogueRequest{
		CloudBackupID:  "badcloudbackup",
		CredentialUUID: ""}).
		Return(&api.BackupCatalogueResponse{CatalogueErr: "Failed to get catalogue"}).Times(1)

	//Invoke Catalogue
	response := client.VolumeDriver(cl).
		BackupCatalogue(&api.BackupCatalogueRequest{
			CloudBackupID:  "goodcloudbackup",
			CredentialUUID: ""})
	require.Equal(t, response.CatalogueErr, "")
	response = client.VolumeDriver(cl).
		BackupCatalogue(&api.BackupCatalogueRequest{
			CloudBackupID:  "badcloudbackup",
			CredentialUUID: ""})
	require.Contains(t, response.CatalogueErr, "Failed to get catalogue")
}

func TestClientBackupHistory(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupHistory(&api.BackupHistoryRequest{
		SrcVolumeID: "goodsrc"}).
		Return(&api.BackupHistoryResponse{HistoryErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupHistory(&api.BackupHistoryRequest{
		SrcVolumeID: "badsrc"}).
		Return(&api.BackupHistoryResponse{HistoryErr: "Failed to get history"}).Times(1)

	//Invoke History
	response := client.VolumeDriver(cl).
		BackupHistory(&api.BackupHistoryRequest{
			SrcVolumeID: "goodsrc"})
	require.Equal(t, response.HistoryErr, "")
	response = client.VolumeDriver(cl).
		BackupHistory(&api.BackupHistoryRequest{
			SrcVolumeID: "badsrc"})
	require.Contains(t, response.HistoryErr, "Failed to get history")
}

func TestClientBackupStateChange(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupStateChange(&api.BackupStateChangeRequest{
		SrcVolumeID:    "goodsrc",
		RequestedState: "pause"}).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupStateChange(&api.BackupStateChangeRequest{
		SrcVolumeID:    "",
		RequestedState: ""}).
		Return(fmt.Errorf("Failed to change state")).Times(1)

	//Invoke StateChange
	err = client.VolumeDriver(cl).
		BackupStateChange(&api.BackupStateChangeRequest{
			SrcVolumeID:    "goodsrc",
			RequestedState: "pause"})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		BackupStateChange(&api.BackupStateChangeRequest{
			SrcVolumeID:    "",
			RequestedState: ""})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Failed to change state")
}

func TestClientBackupSchedCreate(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupSchedCreate(&api.BackupScheduleInfo{
		SrcVolumeID:    "goodsrc",
		CredentialUUID: "",
		BackupSchedule: "daily@10:00"}).
		Return(&api.BackupSchedResponse{SchedCreateErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupSchedCreate(&api.BackupScheduleInfo{
		SrcVolumeID:    "badsrc",
		CredentialUUID: "",
		BackupSchedule: ""}).
		Return(&api.BackupSchedResponse{SchedCreateErr: "Invalid src volume or schedule"}).Times(1)

	//Invoke Create
	response := client.VolumeDriver(cl).
		BackupSchedCreate(&api.BackupScheduleInfo{
			SrcVolumeID:    "goodsrc",
			CredentialUUID: "",
			BackupSchedule: "daily@10:00",
		})
	require.Equal(t, response.SchedCreateErr, "")
	response = client.VolumeDriver(cl).
		BackupSchedCreate(&api.BackupScheduleInfo{
			SrcVolumeID:    "badsrc",
			CredentialUUID: "",
			BackupSchedule: ""})
	require.Contains(t, response.SchedCreateErr, "Invalid src volume or schedule")
}

func TestClientBackupSchedDelete(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupSchedDelete(&api.BackupSchedDeleteRequest{
		SchedUUID: "goodscheduuid"}).
		Return(nil).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupSchedDelete(&api.BackupSchedDeleteRequest{
		SchedUUID: "badscheduuid"}).
		Return(fmt.Errorf("Invalid Schedule UUID")).Times(1)

	//Invoke SchedDelete
	err = client.VolumeDriver(cl).
		BackupSchedDelete(&api.BackupSchedDeleteRequest{
			SchedUUID: "goodscheduuid"})
	require.NoError(t, err)
	err = client.VolumeDriver(cl).
		BackupSchedDelete(&api.BackupSchedDeleteRequest{
			SchedUUID: "badscheduuid"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "Invalid Schedule UUID")
}

func TestClientBackupSchedEnumerate(t *testing.T) {
	ts, testVolDriver := Setup(t)
	defer ts.Close()
	defer testVolDriver.Stop()

	cl, err := client.NewDriverClient(ts.URL, mockDriverName, "", mockDriverName)
	require.NoError(t, err)

	testVolDriver.MockDriver().EXPECT().BackupSchedEnumerate().
		Return(&api.BackupSchedEnumerateResponse{SchedEnumerateErr: ""}).Times(1)
	testVolDriver.MockDriver().EXPECT().BackupSchedEnumerate().
		Return(&api.BackupSchedEnumerateResponse{SchedEnumerateErr: "Failed to Enumerate cloudsnap Schedules"}).Times(1)

	//Invoke Schedule Enumerate
	response := client.VolumeDriver(cl).BackupSchedEnumerate()
	require.Equal(t, response.SchedEnumerateErr, "")
	response = client.VolumeDriver(cl).BackupSchedEnumerate()
	require.Contains(t, response.SchedEnumerateErr, "Failed to Enumerate")
}
