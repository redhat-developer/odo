package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
)

func (vd *volAPI) credsEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	method := "credsEnumerate"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	creds, err := d.CredsEnumerate()
	if err != nil {
		e := fmt.Errorf("Failed to get credential list: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(creds)
}

func (vd *volAPI) credsCreate(w http.ResponseWriter, r *http.Request) {
	var err error
	var input api.CredCreateRequest
	response := &api.CredCreateResponse{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.CredErr = err.Error()
		json.NewEncoder(w).Encode(response)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	response.UUID, err = d.CredsCreate(input.InputParams)
	if err != nil {
		response.CredErr = err.Error()
	}
	json.NewEncoder(w).Encode(response)
}

func (vd *volAPI) credsDelete(w http.ResponseWriter, r *http.Request) {
	var err error
	volumeResponse := &api.VolumeResponse{}

	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]
	if !ok {
		volumeResponse.Error = "Could not parse form for uuid"
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	if err := d.CredsDelete(uuid); err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

func (vd *volAPI) credsValidate(w http.ResponseWriter, r *http.Request) {
	var err error
	volumeResponse := &api.VolumeResponse{}
	vars := mux.Vars(r)
	uuid, ok := vars["uuid"]
	if !ok {
		volumeResponse.Error = "Could not parse form for uuid"
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if err := d.CredsValidate(uuid); err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}
