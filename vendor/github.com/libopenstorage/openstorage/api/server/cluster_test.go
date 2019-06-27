package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	types "github.com/libopenstorage/gossip/types"
	"github.com/libopenstorage/openstorage/api"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"
	mockcluster "github.com/libopenstorage/openstorage/cluster/mock"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"
	"github.com/kubernetes-csi/csi-test/utils"
)

type testCluster struct {
	c       *mockcluster.MockCluster
	mc      *gomock.Controller
	oldInst func() (cluster.Cluster, error)
}

func newTestClutser(t *testing.T) *testCluster {
	tester := &testCluster{}

	// Save already set value of cluster.Inst to set it back
	// when we finish the tests by the defer()
	tester.oldInst = cluster.Inst

	// Create mock controller
	tester.mc = gomock.NewController(&utils.SafeGoroutineTester{})

	// Create a new mock cluster
	tester.c = mockcluster.NewMockCluster(tester.mc)

	// Override cluster.Inst to return our mock cluster
	cluster.Inst = func() (cluster.Cluster, error) {
		return tester.c, nil
	}

	return tester
}

func (c *testCluster) MockCluster() *mockcluster.MockCluster {
	return c.c
}

func (c *testCluster) Finish() {
	cluster.Inst = c.oldInst
	c.mc.Finish()
}
func TestClusterEnumerateSuccess(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.enumerate))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		Enumerate().
		Return(api.Cluster{
			Id:            "cluster-dummy-id",
			Status:        api.Status_STATUS_OK,
			ManagementURL: "mgmturl:1234/mgmt-endpoint",
			Nodes: []api.Node{
				api.Node{
					Hostname: "node1-hostname",
					Id:       "1",
				},
				api.Node{
					Hostname: "node2-hostname",
					Id:       "2",
				},
				api.Node{
					Hostname: "node3-hostname",
					Id:       "3",
				},
			},
		}, nil)
	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.Enumerate()

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	assert.EqualValues(t, "cluster-dummy-id", resp.Id)

}

func TestGossipStateSuccess(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.gossipState))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		GetGossipState().
		Return(&cluster.ClusterState{
			NodeStatus: []types.NodeValue{
				{
					GenNumber: uint64(1234),
					Id:        "node1-id",
					Status:    types.NODE_STATUS_UP,
				},
				{
					GenNumber: uint64(4567),
					Id:        "node2-id",
					Status:    types.NODE_STATUS_UP,
				},
				{
					GenNumber: uint64(7890),
					Id:        "node3-id",
					Status:    types.NODE_STATUS_UP,
				},
			},
		})

		// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.GetGossipState()

	assert.NotNil(t, resp)

	assert.Len(t, resp.NodeStatus, 3)
	assert.EqualValues(t, "node1-id", resp.NodeStatus[0].Id)

}

func TestGossipStateFailed(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.gossipState))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		GetGossipState().
		Return(&cluster.ClusterState{})

		// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.GetGossipState()

	assert.NotNil(t, resp)

	assert.Len(t, resp.NodeStatus, 0)

}
func TestClusterNodeStatusSuccess(t *testing.T) {

	// Create a new global test cluster
	c := newTestClutser(t)
	defer c.Finish()

	// Create an instance of clusterAPI to get access to
	// nodeStatus receiver
	capi := &clusterApi{}

	// Send call to server
	ts := httptest.NewServer(http.HandlerFunc(capi.nodeStatus))
	restClient, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// Set expections
	c.MockCluster().
		EXPECT().
		NodeStatus().
		Return(api.Status_STATUS_OK, nil).
		Times(1)

	// Check status
	status, err := clusterclient.ClusterManager(restClient).NodeStatus()
	assert.NoError(t, err)
	assert.Equal(t, api.Status_STATUS_OK, status)
}

func TestNodeRemoveSuccess(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.delete))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	nodeId := "dummy-node-id-121"
	secondNodeId := "dummy-node-id-131"

	nodes := []api.Node{
		{Id: nodeId},
		{Id: secondNodeId},
	}

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		Remove(nodes, false).
		Return(nil)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.Remove(nodes, false)

	assert.NoError(t, resp)
}

func TestNodeRemoveFailed(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.delete))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	nodeId := ""

	nodes := []api.Node{
		{Id: nodeId},
	}

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		Remove(nodes, false).
		Return(fmt.Errorf("error in removing node"))

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.Remove(nodes, false)

	assert.Error(t, resp)

	assert.Contains(t, resp.Error(), "error in removing node")

}

func TestEnableGossipSuccess(t *testing.T) {
	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.enableGossip))

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		EnableUpdates().
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.EnableUpdates()

	assert.NoError(t, resp)

}

func TestDisableGossipSuccess(t *testing.T) {
	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.disableGossip))

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		DisableUpdates().
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.DisableUpdates()

	assert.NoError(t, resp)

}

func TestSetLoggingURLSuccess(t *testing.T) {

	// Create a new global test cluster
	tc := newTestClutser(t)
	defer tc.Finish()

	// create an instance of clusterAPI to get access to
	// versions endpoint handler

	capi := &clusterApi{}

	// create a HTTP Test server
	ts := httptest.NewServer(http.HandlerFunc(capi.setLoggingURL))

	loggingURL := "http://ip-address:port/dummy-logging-url"

	// mock the cluster response
	tc.MockCluster().
		EXPECT().
		SetLoggingURL(loggingURL).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp := restClient.SetLoggingURL(loggingURL)

	assert.NoError(t, resp)

}
