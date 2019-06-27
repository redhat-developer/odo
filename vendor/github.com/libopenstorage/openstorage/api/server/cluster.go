package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/cluster"
)

const (
	nodeOkMsg    = "Node status OK"
	nodeNotOkMsg = "Node status not OK"
)

type clusterApi struct {
	restBase
}

func (c *clusterApi) Routes() []*Route {
	return []*Route{
		{verb: "GET", path: "/cluster/versions", fn: c.versions},
		{verb: "GET", path: clusterPath("/enumerate", cluster.APIVersion), fn: c.enumerate},
		{verb: "GET", path: clusterPath("/gossipstate", cluster.APIVersion), fn: c.gossipState},
		{verb: "GET", path: clusterPath("/nodestatus", cluster.APIVersion), fn: c.nodeStatus},
		{verb: "GET", path: clusterPath("/nodehealth", cluster.APIVersion), fn: c.nodeHealth},
		{verb: "GET", path: clusterPath("/status", cluster.APIVersion), fn: c.status},
		{verb: "GET", path: clusterPath("/peerstatus", cluster.APIVersion), fn: c.peerStatus},
		{verb: "GET", path: clusterPath("/inspect/{id}", cluster.APIVersion), fn: c.inspect},
		{verb: "DELETE", path: clusterPath("", cluster.APIVersion), fn: c.delete},
		{verb: "DELETE", path: clusterPath("/{id}", cluster.APIVersion), fn: c.delete},
		{verb: "PUT", path: clusterPath("/enablegossip", cluster.APIVersion), fn: c.enableGossip},
		{verb: "PUT", path: clusterPath("/disablegossip", cluster.APIVersion), fn: c.disableGossip},
		{verb: "PUT", path: clusterPath("/shutdown", cluster.APIVersion), fn: c.shutdown},
		{verb: "PUT", path: clusterPath("/shutdown/{id}", cluster.APIVersion), fn: c.shutdown},
		{verb: "GET", path: clusterPath("/alerts/{resource}", cluster.APIVersion), fn: c.enumerateAlerts},
		{verb: "PUT", path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.clearAlert},
		{verb: "DELETE", path: clusterPath("/alerts/{resource}/{id}", cluster.APIVersion), fn: c.eraseAlert},
	}
}

func newClusterAPI() restServer {
	return &clusterApi{restBase{version: cluster.APIVersion, name: "Cluster API"}}
}

func (c *clusterApi) String() string {
	return c.name
}

// swagger:operation GET /cluster/enumerate cluster enumerate enumerateCluster
//
// Lists cluster Nodes.
//
// This will return the entire cluster object and it's nodes.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: current cluster state
//      schema:
//         type: array
//         items:
//            $ref: '#/definitions/Cluster'
func (c *clusterApi) enumerate(w http.ResponseWriter, r *http.Request) {
	method := "enumerate"
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	cluster, err := inst.Enumerate()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cluster)
}

func (c *clusterApi) setSize(w http.ResponseWriter, r *http.Request) {
	method := "set size"
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := r.URL.Query()

	size := params["size"]
	if size == nil {
		c.sendError(c.name, method, w, "Missing size param", http.StatusBadRequest)
		return
	}

	sz, _ := strconv.Atoi(size[0])

	err = inst.SetSize(sz)

	clusterResponse := &api.ClusterResponse{Error: err.Error()}
	json.NewEncoder(w).Encode(clusterResponse)
}

// swagger:operation GET /cluster/inspect/{id} cluster inspect inspectNode
//
// Inspect cluster Nodes.
//
// This will return the requested node object
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
// responses:
//   '200':
//      description: a node
//      schema:
//       $ref: '#/definitions/Node'
func (c *clusterApi) inspect(w http.ResponseWriter, r *http.Request) {
	method := "inspect"
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	nodeID, ok := vars["id"]

	if !ok || nodeID == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	if nodeStats, err := inst.Inspect(nodeID); err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
	} else {
		json.NewEncoder(w).Encode(nodeStats)
	}
}

// swagger:operation PUT /loggingurl cluster loggingurl setLoggingUrl
//
// Set Logging url
// ---
// produces:
// - application/json
// deprecated: true
// parameters:
// - name: url
//   in: query
//   description: url to set loggingurl with
//   required: true
// responses:
//  '200':
//    description: cluster response
//    schema:
//     $ref: '#/definitions/ClusterResponse'
func (c *clusterApi) setLoggingURL(w http.ResponseWriter, r *http.Request) {
	method := "set Logging URL"

	inst, err := cluster.Inst()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := r.URL.Query()
	loggingURL := params["url"]
	if len(loggingURL) == 0 {
		c.sendError(c.name, method, w, "Missing url param  url", http.StatusBadRequest)
		return
	}

	err = inst.SetLoggingURL(strings.TrimSpace(loggingURL[0]))

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&api.ClusterResponse{})
}

func (c *clusterApi) enableGossip(w http.ResponseWriter, r *http.Request) {
	method := "enablegossip"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.EnableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) disableGossip(w http.ResponseWriter, r *http.Request) {
	method := "disablegossip"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	inst.DisableUpdates()

	clusterResponse := &api.ClusterResponse{}
	json.NewEncoder(w).Encode(clusterResponse)
}

func (c *clusterApi) gossipState(w http.ResponseWriter, r *http.Request) {
	method := "gossipState"

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := inst.GetGossipState()
	json.NewEncoder(w).Encode(resp)
}

// swagger:operation GET /cluster/status cluster status status
//
// this will return the cluster status.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: cluster status
//      schema:
//         type: string
func (c *clusterApi) status(w http.ResponseWriter, r *http.Request) {
	method := "status"

	inst, err := cluster.Inst()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	cluster, err := inst.Enumerate()

	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cluster.Status)
}

func nodeStatusIntl() (api.Status, error) {
	inst, err := cluster.Inst()
	if err != nil {
		return api.Status_STATUS_NONE, err
	}

	resp, err := inst.NodeStatus()
	if err != nil {
		return api.Status_STATUS_NONE, err
	}

	return resp, nil
}

// swagger:operation GET /cluster/nodestatus node status nodeStatus
//
// This will return the node status .
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node status of responding node.
//      schema:
//         type: string
func (c *clusterApi) nodeStatus(w http.ResponseWriter, r *http.Request) {
	method := "nodeStatus"

	st, err := nodeStatusIntl()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(st)
}

// swagger:operation GET /cluster/nodehealth node health nodeHealth
//
// This will return node health.
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: node health of responding node.
//      schema:
//         type: string
func (c *clusterApi) nodeHealth(w http.ResponseWriter, r *http.Request) {
	method := "nodeHealth"

	st, err := nodeStatusIntl()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if st != api.Status_STATUS_OK {
		err = fmt.Errorf("%s (%s)", nodeNotOkMsg, api.Status_name[int32(st)])
		c.sendError(c.name, method, w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Write([]byte(nodeOkMsg + "\n"))
}

// swagger:operation GET /cluster/peerstatus node peerstatus peerStatus
//
// This will return the peer node status
//
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: path
//   description: id of the node we want to check.
//   required: true
// responses:
//   '200':
//      description: node status of requested node
//      schema:
//         type: string
func (c *clusterApi) peerStatus(w http.ResponseWriter, r *http.Request) {
	method := "peerStatus"

	params := r.URL.Query()
	listenerName := params["name"]
	if len(listenerName) == 0 || listenerName[0] == "" {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}
	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp, err := inst.PeerStatus(listenerName[0])
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)

}

// swagger:operation DELETE /cluster/{id} cluster node delete deleteNode
//
// This will delete a node from the cluster
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
// - name: forceRemove
//   in: query
//   description: forceRemove node
//   required: false
// responses:
//   '200':
//      description: delete node success
//      schema:
//         type: string
func (c *clusterApi) delete(w http.ResponseWriter, r *http.Request) {
	method := "delete"

	params := r.URL.Query()

	nodeID := params["id"]
	if nodeID == nil {
		c.sendError(c.name, method, w, "Missing id param", http.StatusBadRequest)
		return
	}

	forceRemoveParam := params["forceRemove"]
	forceRemove := false
	if forceRemoveParam != nil {
		var err error
		forceRemove, err = strconv.ParseBool(forceRemoveParam[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid forceRemove Option: "+
				forceRemoveParam[0], http.StatusBadRequest)
			return
		}
	}

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	nodes := make([]api.Node, 0)
	for _, id := range nodeID {
		nodes = append(nodes, api.Node{Id: id})
	}

	clusterResponse := &api.ClusterResponse{}

	err = inst.Remove(nodes, forceRemove)
	if err != nil {
		clusterResponse.Error = fmt.Errorf("Node Remove: %s", err).Error()
	}
	json.NewEncoder(w).Encode(clusterResponse)
}

// swagger:operation PUT /cluster/{id} cluster node shutdown shutdownNode
//
// This will shutdown a node (Not Implemented)
//
// ---
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: id to get node with
//   required: true
// responses:
//   '200':
//      description: shutdown success
//      schema:
//         type: string
func (c *clusterApi) shutdown(w http.ResponseWriter, r *http.Request) {
	method := "shutdown"
	c.sendNotImplemented(w, method)
}

// swagger:operation GET /cluster/versions cluster versions enumerateVersions
//
// Lists API Versions supported by this cluster
//
// ---
// produces:
// - application/json
// responses:
//   '200':
//      description: anode
//      schema:
//         type: array
//         items:
//            type: string
func (c *clusterApi) versions(w http.ResponseWriter, r *http.Request) {
	versions := []string{
		cluster.APIVersion,
		// Update supported versions by adding them here
	}
	json.NewEncoder(w).Encode(versions)
}

// swagger:operation GET /cluster/alerts/{resource} cluster alerts enumerate enumerateAlerts
//
// This will return a list of alerts for the requested resource
//
// ---
// produces:
// - application/json
// parameters:
// - name: resource
//   in: path
//   description: resourcetype to get alerts with
//   required: true
//   schema:
//    "$ref": "#/definitions/ResourceType"
// responses:
//   '200':
//      description: Alerts object
//      schema:
//       $ref: '#/definitions/Alerts'
func (c *clusterApi) enumerateAlerts(w http.ResponseWriter, r *http.Request) {
	method := "enumerateAlerts"

	params := r.URL.Query()

	var (
		resourceType api.ResourceType
		err          error
		tS, tE       time.Time
	)
	vars := mux.Vars(r)
	resource, ok := vars["resource"]
	if ok {
		resourceType, err = handleResourceType(resource)
		if err != nil {
			c.sendError(c.name, method, w, "Invalid resource param", http.StatusBadRequest)
			return
		}
	} else {
		resourceType = api.ResourceType_RESOURCE_TYPE_NONE
	}

	timeStart := params["timestart"]
	if timeStart != nil {
		tS, err = time.Parse(api.TimeLayout, timeStart[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid timestart param", http.StatusBadRequest)
			return
		}
	}

	timeEnd := params["timeend"]
	if timeEnd != nil {
		tS, err = time.Parse(api.TimeLayout, timeEnd[0])
		if err != nil {
			c.sendError(c.name, method, w, "Invalid timeend param", http.StatusBadRequest)
			return
		}
	}

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	alerts, err := inst.EnumerateAlerts(tS, tE, resourceType)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(alerts)
}

// swagger:operation PUT /cluster/alerts/{resource}/{id} cluster alerts clear clearAlert
//
// This will clear alert {id} with resourcetype {resource}
//
// ---
// produces:
// - application/json
// parameters:
// - name: resource
//   in: path
//   description: resourcetype to get alerts with
//   required: true
//   schema:
//    "$ref": "#/definitions/ResourceType"
// - name: id
//   in: path
//   description: id to get alerts with
//   required: true
// responses:
//   '200':
//      description: Alerts object
//      schema:
//       type: string
func (c *clusterApi) clearAlert(w http.ResponseWriter, r *http.Request) {
	method := "clearAlert"

	resourceType, alertId, err := c.getAlertParams(w, r, method)
	if err != nil {
		return
	}

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.ClearAlert(resourceType, alertId)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode("Successfully cleared Alert")
}

// swagger:operation DELETE /cluster/alerts/{resource}/{id} cluster alerts delete deleteAlert
//
// This delete clear alert {id} with resourcetype {resource}
//
// ---
// produces:
// - application/json
// parameters:
// - name: resource
//   in: path
//   description: resourcetype to get alerts with
//   required: true
//   schema:
//    "$ref": "#/definitions/ResourceType"
// - name: id
//   in: path
//   description: id to get alerts with
//   required: true
// responses:
//   '200':
//      description: Alerts object
//      schema:
//       type: string
func (c *clusterApi) eraseAlert(w http.ResponseWriter, r *http.Request) {
	method := "eraseAlert"

	resourceType, alertId, err := c.getAlertParams(w, r, method)
	if err != nil {
		return
	}

	inst, err := cluster.Inst()
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = inst.EraseAlert(resourceType, alertId)
	if err != nil {
		c.sendError(c.name, method, w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode("Successfully erased Alert")
}

func (c *clusterApi) getAlertParams(w http.ResponseWriter, r *http.Request, method string) (api.ResourceType, int64, error) {
	var (
		resourceType api.ResourceType
		alertId      int64
		err          error
	)
	returnErr := fmt.Errorf("Invalid param")

	vars := mux.Vars(r)
	resource, ok := vars["resource"]
	if ok {
		resourceType, err = handleResourceType(resource)
	}

	if err != nil || !ok {
		c.sendError(c.name, method, w, "Missing/Invalid resource param", http.StatusBadRequest)
		return api.ResourceType_RESOURCE_TYPE_NONE, 0, returnErr

	}

	vars = mux.Vars(r)
	id, ok := vars["id"]
	if ok {
		alertId, err = strconv.ParseInt(id, 10, 64)
	}

	if err != nil || !ok {
		c.sendError(c.name, method, w, "Missing/Invalid id param", http.StatusBadRequest)
		return api.ResourceType_RESOURCE_TYPE_NONE, 0, returnErr
	}
	return resourceType, alertId, nil
}

func (c *clusterApi) sendNotImplemented(w http.ResponseWriter, method string) {
	c.sendError(c.name, method, w, "Not implemented.", http.StatusNotImplemented)
}

func clusterVersion(route, version string) string {
	return "/" + version + "/" + route
}

func clusterPath(route, version string) string {
	return clusterVersion("cluster"+route, version)
}

func handleResourceType(resource string) (api.ResourceType, error) {
	resource = strings.ToLower(resource)
	switch resource {
	case "volume":
		return api.ResourceType_RESOURCE_TYPE_VOLUME, nil
	case "node":
		return api.ResourceType_RESOURCE_TYPE_NODE, nil
	case "cluster":
		return api.ResourceType_RESOURCE_TYPE_CLUSTER, nil
	case "drive":
		return api.ResourceType_RESOURCE_TYPE_DRIVE, nil
	default:
		resourceType, err := strconv.ParseInt(resource, 10, 64)
		if err == nil {
			if _, ok := api.ResourceType_name[int32(resourceType)]; ok {
				return api.ResourceType(resourceType), nil
			}
		}
		return api.ResourceType_RESOURCE_TYPE_NONE, fmt.Errorf("Invalid resource type")
	}
}
