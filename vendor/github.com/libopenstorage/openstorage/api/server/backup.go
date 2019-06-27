package server

import (
	"encoding/json"
	"github.com/libopenstorage/openstorage/api"
	"net/http"
)

func (vd *volAPI) backup(w http.ResponseWriter, r *http.Request) {
	var err error
	backupReq := &api.BackupRequest{}
	method := "backup"

	if err := json.NewDecoder(r.Body).Decode(backupReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	volumeResponse := &api.VolumeResponse{}
	err = d.Backup(backupReq)
	if err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) backuprestore(w http.ResponseWriter, r *http.Request) {
	restoreReq := &api.BackupRestoreRequest{}
	restoreResp := &api.BackupRestoreResponse{}

	if err := json.NewDecoder(r.Body).Decode(restoreReq); err != nil {
		restoreResp.RestoreErr = err.Error()
		json.NewEncoder(w).Encode(restoreResp)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	restoreResp = d.BackupRestore(restoreReq)
	json.NewEncoder(w).Encode(restoreResp)
}

func (vd *volAPI) backupdelete(w http.ResponseWriter, r *http.Request) {
	var err error
	backupReq := &api.BackupDeleteRequest{}
	method := "backupdelete"

	if err := json.NewDecoder(r.Body).Decode(backupReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	volumeResponse := &api.VolumeResponse{}
	err = d.BackupDelete(backupReq)
	if err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) backupenumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	enumerateReq := &api.BackupEnumerateRequest{}
	enumerateResp := &api.BackupEnumerateResponse{}

	if err := json.NewDecoder(r.Body).Decode(enumerateReq); err != nil {
		enumerateResp.EnumerateErr = err.Error()
		json.NewEncoder(w).Encode(enumerateResp)
		return
	}
	d, err := vd.getVolDriver(r)

	if err != nil {
		notFound(w, r)
		return
	}
	enumerateResp = d.BackupEnumerate(enumerateReq)

	json.NewEncoder(w).Encode(enumerateResp)
}

func (vd *volAPI) backupstatus(w http.ResponseWriter, r *http.Request) {
	var err error
	backupSts := &api.BackupStsRequest{}
	backupStsResp := &api.BackupStsResponse{}

	if err := json.NewDecoder(r.Body).Decode(backupSts); err != nil {
		backupStsResp.StsErr = err.Error()
		json.NewEncoder(w).Encode(backupStsResp)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	backupStsResp = d.BackupStatus(backupSts)
	if err != nil {
		backupStsResp.StsErr = err.Error()
	}
	json.NewEncoder(w).Encode(backupStsResp)
}

func (vd *volAPI) backupcatalogue(w http.ResponseWriter, r *http.Request) {
	var err error
	catalogueReq := &api.BackupCatalogueRequest{}
	catalogue := &api.BackupCatalogueResponse{}

	if err := json.NewDecoder(r.Body).Decode(catalogueReq); err != nil {
		catalogue.CatalogueErr = err.Error()
		json.NewEncoder(w).Encode(catalogue)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	catalogue = d.BackupCatalogue(catalogueReq)
	json.NewEncoder(w).Encode(catalogue)

}
func (vd *volAPI) backuphistory(w http.ResponseWriter, r *http.Request) {
	var err error
	historyReq := &api.BackupHistoryRequest{}
	history := &api.BackupHistoryResponse{}

	if err := json.NewDecoder(r.Body).Decode(historyReq); err != nil {
		history.HistoryErr = err.Error()
		json.NewEncoder(w).Encode(history)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	history = d.BackupHistory(historyReq)
	json.NewEncoder(w).Encode(history)
}

func (vd *volAPI) backupstatechange(w http.ResponseWriter, r *http.Request) {
	var err error
	stateChangeReq := &api.BackupStateChangeRequest{}
	volumeResponse := &api.VolumeResponse{}
	if err := json.NewDecoder(r.Body).Decode(stateChangeReq); err != nil {
		volumeResponse.Error = responseStatus(err)
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	err = d.BackupStateChange(stateChangeReq)
	if err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) backupschedcreate(w http.ResponseWriter, r *http.Request) {
	var err error
	backupSchedReq := &api.BackupScheduleInfo{}
	backupSchedResp := &api.BackupSchedResponse{}
	if err := json.NewDecoder(r.Body).Decode(backupSchedReq); err != nil {
		backupSchedResp.SchedCreateErr = err.Error()
		json.NewEncoder(w).Encode(backupSchedResp)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	backupSchedResp = d.BackupSchedCreate(backupSchedReq)
	json.NewEncoder(w).Encode(backupSchedResp)
}

func (vd *volAPI) backupscheddelete(w http.ResponseWriter, r *http.Request) {
	var err error
	deleteReq := &api.BackupSchedDeleteRequest{}
	volumeResponse := &api.VolumeResponse{}
	if err := json.NewDecoder(r.Body).Decode(deleteReq); err != nil {
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	err = d.BackupSchedDelete(deleteReq)
	if err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) backupschedenumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	schedules := &api.BackupSchedEnumerateResponse{}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	schedules = d.BackupSchedEnumerate()
	json.NewEncoder(w).Encode(schedules)
}
