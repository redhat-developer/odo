package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/errors"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"
)

const schedDriverPostFix = "-sched"

type volAPI struct {
	restBase
}

func responseStatus(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func newVolumeAPI(name string) restServer {
	return &volAPI{restBase{version: volume.APIVersion, name: name}}
}

func (vd *volAPI) String() string {
	return vd.name
}

func (vd *volAPI) getVolDriver(r *http.Request) (volume.VolumeDriver, error) {
	// Check if the driver has registered by it's user agent name
	userAgent := r.Header.Get("User-Agent")
	if len(userAgent) > 0 {
		clientName := strings.Split(userAgent, "/")
		if len(clientName) > 0 {
			d, err := volumedrivers.Get(clientName[0])
			if err == nil {
				return d, nil
			}
		}
	}

	// Check if the driver has registered a scheduler-based driver
	d, err := volumedrivers.Get(vd.name + schedDriverPostFix)
	if err == nil {
		return d, nil
	}

	// default
	return volumedrivers.Get(vd.name)
}

func (vd *volAPI) parseID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	if id, ok := vars["id"]; ok {
		return string(id), nil
	}
	return "", fmt.Errorf("could not parse snap ID")
}

// swagger:operation POST /osd-volumes volume create createVolume
//
// Creates a single volume with given spec.
//
// ---
// produces:
// - application/json
// parameters:
// - name: spec
//   in: body
//   description: spec to create volume with
//   required: true
//   schema:
//         "$ref": "#/definitions/VolumeCreateRequest"
// responses:
//   '200':
//     description: volume create response
//     schema:
//         "$ref": "#/definitions/VolumeCreateResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeCreateResponse"

func (vd *volAPI) create(w http.ResponseWriter, r *http.Request) {
	var dcRes api.VolumeCreateResponse
	var dcReq api.VolumeCreateRequest
	method := "create"

	if err := json.NewDecoder(r.Body).Decode(&dcReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	id, err := d.Create(dcReq.Locator, dcReq.Source, dcReq.Spec)
	dcRes.VolumeResponse = &api.VolumeResponse{Error: responseStatus(err)}
	dcRes.Id = id

	vd.logRequest(method, id).Infoln("")

	json.NewEncoder(w).Encode(&dcRes)
}

func processErrorForVolSetResponse(action *api.VolumeStateAction, err error, resp *api.VolumeSetResponse) {
	if err == nil || resp == nil {
		return
	}

	if action != nil && (action.Mount == api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF ||
		action.Attach == api.VolumeActionParam_VOLUME_ACTION_PARAM_OFF) {
		switch err.(type) {
		case *errors.ErrNotFound:
			resp.VolumeResponse = &api.VolumeResponse{}
			resp.Volume = &api.Volume{}
		default:
			resp.VolumeResponse = &api.VolumeResponse{
				Error: err.Error(),
			}
		}
	} else if err != nil {
		resp.VolumeResponse = &api.VolumeResponse{
			Error: err.Error(),
		}
	}
}

// swagger:operation PUT /osd-volumes/{id} volume update setVolume
//
// Updates a single volume with given spec.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: id to get volume with
//   required: true
// - name: spec
//   in: body
//   description: spec to set volume with
//   required: true
//   schema:
//         "$ref": "#/definitions/VolumeSetRequest"
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeSetResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeSetResponse"
func (vd *volAPI) volumeSet(w http.ResponseWriter, r *http.Request) {
	var (
		volumeID string
		err      error
		req      api.VolumeSetRequest
		resp     api.VolumeSetResponse
	)
	method := "volumeSet"

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	if volumeID, err = vd.parseID(r); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}

	setActions := ""
	if req.Action != nil {
		setActions = fmt.Sprintf("Mount=%v Attach=%v", req.Action.Mount, req.Action.Attach)
	}

	vd.logRequest(method, string(volumeID)).Infoln(setActions)

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if req.Locator != nil || req.Spec != nil {
		err = d.Set(volumeID, req.Locator, req.Spec)
	}

	for err == nil && req.Action != nil {
		if req.Action.Attach != api.VolumeActionParam_VOLUME_ACTION_PARAM_NONE {
			if req.Action.Attach == api.VolumeActionParam_VOLUME_ACTION_PARAM_ON {
				_, err = d.Attach(volumeID, req.Options)
			} else {
				err = d.Detach(volumeID, req.Options)
			}
			if err != nil {
				break
			}
		}

		if req.Action.Mount != api.VolumeActionParam_VOLUME_ACTION_PARAM_NONE {
			if req.Action.Mount == api.VolumeActionParam_VOLUME_ACTION_PARAM_ON {
				if req.Action.MountPath == "" {
					err = fmt.Errorf("Invalid mount path")
					break
				}
				err = d.Mount(volumeID, req.Action.MountPath, req.Options)
			} else {
				err = d.Unmount(volumeID, req.Action.MountPath, req.Options)
			}
			if err != nil {
				break
			}
		}
		break
	}

	if err != nil {
		processErrorForVolSetResponse(req.Action, err, &resp)
	} else {
		v, err := d.Inspect([]string{volumeID})
		if err != nil {
			processErrorForVolSetResponse(req.Action, err, &resp)
		} else if v == nil || len(v) != 1 {
			processErrorForVolSetResponse(req.Action, &errors.ErrNotFound{Type: "Volume", ID: volumeID}, &resp)
		} else {
			v0 := v[0]
			resp.Volume = v0
		}
	}

	json.NewEncoder(w).Encode(resp)

}

// swagger:operation GET /osd-volumes/{id} volume inspect inspectVolume
//
// Inspect volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume get response
//     schema:
//         "$ref": "#/definitions/Volume"
func (vd *volAPI) inspect(w http.ResponseWriter, r *http.Request) {
	var err error
	var volumeID string

	method := "inspect"
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	dk, err := d.Inspect([]string{volumeID})
	if err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

// swagger:operation DELETE /osd-volumes/{id} volume delete deleteVolume
//
// Delete volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) delete(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	method := "delete"
	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	vd.logRequest(method, volumeID).Infoln("")

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	volumeResponse := &api.VolumeResponse{}

	if err := d.Delete(volumeID); err != nil {
		volumeResponse.Error = err.Error()
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation GET /osd-volumes volume enumerate enumerateVolumes
//
// Enumerate all volumes
//
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: User specified volume name (Case Sensitive)
//   required: false
//   type: string
// - name: Label
//   in: formData
//   description: Comma separated name value pairs
//   required: false
//   schema:
//    type: object
//	  example: {"label1","label2"} # Example value
// - name: ConfigLabel
//   in: formData
//   description: Comma separated name value pairs
//   required: false
//   schema:
//    type: object
//	  example: {"label1","label2"} # Example value
// - name: VolumeID
//   in: query
//   description: Volume UUID
//   required: false
//   type: string
//   format: uuid
// responses:
//   '200':
//      description: an array of volumes
//      schema:
//         type: array
//         items:
//            $ref: '#/definitions/Volume'
func (vd *volAPI) enumerate(w http.ResponseWriter, r *http.Request) {
	var locator api.VolumeLocator
	var configLabels map[string]string
	var err error
	var vols []*api.Volume

	method := "enumerate"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	params := r.URL.Query()
	v := params[string(api.OptName)]
	if v != nil {
		locator.Name = v[0]
	}
	v = params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &locator.VolumeLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}
	v = params[string(api.OptConfigLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &configLabels); err != nil {
			e := fmt.Errorf("Failed to parse parse configLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}
	v = params[string(api.OptVolumeID)]
	if v != nil {
		ids := make([]string, len(v))
		for i, s := range v {
			ids[i] = string(s)
		}
		vols, err = d.Inspect(ids)
		if err != nil {
			e := fmt.Errorf("Failed to inspect volumeID: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
			return
		}
	} else {
		vols, err = d.Enumerate(&locator, configLabels)
		if err != nil {
			vd.sendError(vd.name, method, w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	json.NewEncoder(w).Encode(vols)
}

// swagger:operation POST /osd-snapshots snapshot create createSnap
//
// Take a snapshot of volume in SnapCreateRequest
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: id to get volume with
//   required: true
// - name: spec
//   in: body
//   description: spec to create snap with
//   required: true
//   schema:
//    "$ref": "#/definitions/SnapCreateRequest"
// responses:
//    '200':
//      description: an array of volumes
//      schema:
//       "$ref": '#/definitions/SnapCreateResponse'
//    default:
//     description: unexpected error
//     schema:
//      "$ref": "#/definitions/SnapCreateResponse"
func (vd *volAPI) snap(w http.ResponseWriter, r *http.Request) {
	var snapReq api.SnapCreateRequest
	var snapRes api.SnapCreateResponse
	method := "snap"

	if err := json.NewDecoder(r.Body).Decode(&snapReq); err != nil {
		vd.sendError(vd.name, method, w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	vd.logRequest(method, string(snapReq.Id)).Infoln("")

	id, err := d.Snapshot(snapReq.Id, snapReq.Readonly, snapReq.Locator)
	snapRes.VolumeCreateResponse = &api.VolumeCreateResponse{
		Id: id,
		VolumeResponse: &api.VolumeResponse{
			Error: responseStatus(err),
		},
	}
	json.NewEncoder(w).Encode(&snapRes)
}

// swagger:operation POST /osd-snapshots/restore/{id} snapshot restore restoreSnap
//
// Restore snapshot with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id of snapshot to restore
//   required: true
// responses:
//  '200':
//    description: Restored volume
//    schema:
//     "$ref": '#/definitions/VolumeResponse'
//  default:
//   description: unexpected error
//   schema:
//    "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) restore(w http.ResponseWriter, r *http.Request) {
	var volumeID, snapID string
	var err error
	method := "restore"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	params := r.URL.Query()
	v := params[api.OptSnapID]
	if v != nil {
		snapID = v[0]
	} else {
		vd.sendError(vd.name, method, w, "Missing "+api.OptSnapID+" param",
			http.StatusBadRequest)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	if err := d.Restore(volumeID, snapID); err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation GET /osd-snapshots snapshot enumerate enumerateSnaps
//
// Enumerate snapshots.
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: query
//   description: Volume name that maps to this snap
//   required: false
//   type: string
// - name: VolumeLabels
//   in: formData
//   description: Comma separated volume labels
//   required: false
//   schema:
//    type: object
//	  example: {"label1","label2"} # Example value
// - name: SnapLabels
//   in: formData
//   description: Comma separated snap labels
//   required: false
//   schema:
//    type: object
//	  example: {"label1","label2"} # Example value
// - name: uuid
//   in: query
//   description: Snap UUID
//   required: false
//   type: string
//   format: uuid
// responses:
//  '200':
//   description: an array of snapshots
//   schema:
//    type: array
//    items:
//     $ref: '#/definitions/Volume'
func (vd *volAPI) snapEnumerate(w http.ResponseWriter, r *http.Request) {
	var err error
	var labels map[string]string
	var ids []string

	method := "snapEnumerate"
	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}
	params := r.URL.Query()
	v := params[string(api.OptLabel)]
	if v != nil {
		if err = json.Unmarshal([]byte(v[0]), &labels); err != nil {
			e := fmt.Errorf("Failed to parse parse VolumeLabels: %s", err.Error())
			vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		}
	}

	v, ok := params[string(api.OptVolumeID)]
	if v != nil && ok {
		ids = make([]string, len(params))
		for i, s := range v {
			ids[i] = string(s)
		}
	}

	snaps, err := d.SnapEnumerate(ids, labels)
	if err != nil {
		e := fmt.Errorf("Failed to enumerate snaps: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(snaps)
}

// swagger:operation GET /osd-volumes/stats/{id} volume stats statsVolume
//
// Get stats for volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//  '200':
//   description: volume set response
//   schema:
//    "$ref": "#/definitions/Stats"
func (vd *volAPI) stats(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	// By default always report /proc/diskstats style stats.
	cumulative := true
	if opt, ok := params[string(api.OptCumulative)]; ok {
		if boolValue, err := strconv.ParseBool(strings.Join(opt[:], "")); !ok {
			e := fmt.Errorf("Failed to parse %s option: %s",
				api.OptCumulative, err.Error())
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		} else {
			cumulative = boolValue
		}
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	stats, err := d.Stats(volumeID, cumulative)
	if err != nil {
		e := fmt.Errorf("Failed to get stats: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(stats)
}

// swagger:operation GET /osd-volumes/usedsize/{id} volume usedsize usedSizeVolume
//
// Get Used size of volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//  '200':
//   description: volume set response
//   type: integer
//   format: int64
func (vd *volAPI) usedsize(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse volumeID: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	used, err := d.UsedSize(volumeID)
	if err != nil {
		e := fmt.Errorf("Failed to get used size: %s", err.Error())
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(used)
}

// swagger:operation POST /osd-volumes/requests/{id} volume requests requestsVolume
//
// Get Requests for volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/ActiveRequests"
func (vd *volAPI) requests(w http.ResponseWriter, r *http.Request) {
	var err error

	method := "requests"

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	requests, err := d.GetActiveRequests()
	if err != nil {
		e := fmt.Errorf("Failed to get active requests: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(requests)
}

// swagger:operation GET /osd-volumes/quiesce/{id} volume quiesce quiesceVolume
//
// Quiesce volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) quiesce(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error
	method := "quiesce"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	params := r.URL.Query()
	timeoutStr := params[api.OptTimeoutSec]
	var timeoutSec uint64
	if timeoutStr != nil {
		var err error
		timeoutSec, err = strconv.ParseUint(timeoutStr[0], 10, 64)
		if err != nil {
			vd.sendError(vd.name, method, w, api.OptTimeoutSec+" must be int",
				http.StatusBadRequest)
			return
		}
	}

	quiesceIdParam := params[api.OptQuiesceID]
	var quiesceId string
	if len(quiesceIdParam) > 0 {
		quiesceId = quiesceIdParam[0]
	}

	volumeResponse := &api.VolumeResponse{}
	if err := d.Quiesce(volumeID, timeoutSec, quiesceId); err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:operation POST /osd-volumes/unquiesce/{id} volume unquiesce unquiesceVolume
//
// Unquiesce volume with specified id.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume set response
//     schema:
//         "$ref": "#/definitions/VolumeResponse"
//   default:
//     description: unexpected error
//     schema:
//       "$ref": "#/definitions/VolumeResponse"
func (vd *volAPI) unquiesce(w http.ResponseWriter, r *http.Request) {
	var volumeID string
	var err error
	method := "unquiesce"

	if volumeID, err = vd.parseID(r); err != nil {
		e := fmt.Errorf("Failed to parse parse volumeID: %s", err.Error())
		vd.sendError(vd.name, method, w, e.Error(), http.StatusBadRequest)
		return
	}

	d, err := vd.getVolDriver(r)
	if err != nil {
		notFound(w, r)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	if err := d.Unquiesce(volumeID); err != nil {
		volumeResponse.Error = responseStatus(err)
	}
	json.NewEncoder(w).Encode(volumeResponse)
}

// swagger:route GET /osd-volumes/versions volume versions listVersions
//
// Lists API versions supported by this volumeDriver.
//
// This will show all supported versions of the API for this volumeDriver.
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: query
//   description: id to get volume with
//   required: true
// responses:
//   '200':
//     description: volume set response
//	   type: array
//      items: string
func (vd *volAPI) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		volume.APIVersion,
		// Update supported versions by adding them here
	}
	json.NewEncoder(w).Encode(versions)
}

func volVersion(route, version string) string {
	if version == "" {
		return "/" + route
	} else {
		return "/" + version + "/" + route
	}
}

func volPath(route, version string) string {
	return volVersion(api.OsdVolumePath+route, version)
}

func snapPath(route, version string) string {
	return volVersion(api.OsdSnapshotPath+route, version)
}

func credsPath(route, version string) string {
	return volVersion(api.OsdCredsPath+route, version)
}

func backupPath(route, version string) string {
	return volVersion(api.OsdBackupPath+route, version)
}

func (vd *volAPI) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: "/" + api.OsdVolumePath + "/versions", fn: vd.versions},
		{verb: "POST", path: volPath("", volume.APIVersion), fn: vd.create},
		{verb: "PUT", path: volPath("/{id}", volume.APIVersion), fn: vd.volumeSet},
		{verb: "GET", path: volPath("", volume.APIVersion), fn: vd.enumerate},
		{verb: "GET", path: volPath("/{id}", volume.APIVersion), fn: vd.inspect},
		{verb: "DELETE", path: volPath("/{id}", volume.APIVersion), fn: vd.delete},
		{verb: "GET", path: volPath("/stats", volume.APIVersion), fn: vd.stats},
		{verb: "GET", path: volPath("/stats/{id}", volume.APIVersion), fn: vd.stats},
		{verb: "GET", path: volPath("/usedsize", volume.APIVersion), fn: vd.usedsize},
		{verb: "GET", path: volPath("/usedsize/{id}", volume.APIVersion), fn: vd.usedsize},
		{verb: "GET", path: volPath("/requests", volume.APIVersion), fn: vd.requests},
		{verb: "GET", path: volPath("/requests/{id}", volume.APIVersion), fn: vd.requests},
		{verb: "POST", path: volPath("/quiesce/{id}", volume.APIVersion), fn: vd.quiesce},
		{verb: "POST", path: volPath("/unquiesce/{id}", volume.APIVersion), fn: vd.unquiesce},
		{verb: "POST", path: snapPath("", volume.APIVersion), fn: vd.snap},
		{verb: "GET", path: snapPath("", volume.APIVersion), fn: vd.snapEnumerate},
		{verb: "POST", path: snapPath("/restore/{id}", volume.APIVersion), fn: vd.restore},
		{verb: "GET", path: credsPath("", volume.APIVersion), fn: vd.credsEnumerate},
		{verb: "POST", path: credsPath("", volume.APIVersion), fn: vd.credsCreate},
		{verb: "DELETE", path: credsPath("/{uuid}", volume.APIVersion), fn: vd.credsDelete},
		{verb: "POST", path: credsPath("/validate/{uuid}", volume.APIVersion), fn: vd.credsValidate},
		{verb: "POST", path: backupPath("", volume.APIVersion), fn: vd.backup},
		{verb: "POST", path: backupPath("/restore", volume.APIVersion), fn: vd.backuprestore},
		{verb: "GET", path: backupPath("", volume.APIVersion), fn: vd.backupenumerate},
		{verb: "DELETE", path: backupPath("", volume.APIVersion), fn: vd.backupdelete},
		{verb: "POST", path: backupPath("/status", volume.APIVersion), fn: vd.backupstatus},
		{verb: "GET", path: backupPath("/catalogue", volume.APIVersion), fn: vd.backupcatalogue},
		{verb: "GET", path: backupPath("/history", volume.APIVersion), fn: vd.backuphistory},
		{verb: "POST", path: backupPath("/statechange", volume.APIVersion), fn: vd.backupstatechange},
		{verb: "POST", path: backupPath("/schedcreate", volume.APIVersion), fn: vd.backupschedcreate},
		{verb: "POST", path: backupPath("/scheddelete", volume.APIVersion), fn: vd.backupscheddelete},
		{verb: "GET", path: backupPath("/schedenumerate", volume.APIVersion), fn: vd.backupschedenumerate},
	}
}
