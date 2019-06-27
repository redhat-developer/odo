// +build daemon

/*
CSI Interface for OSD
Copyright 2017 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package csi

import (
	"os"
	"testing"

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
	"github.com/libopenstorage/openstorage/volume/drivers"

	"github.com/kubernetes-csi/csi-test/pkg/sanity"
	"go.pedge.io/dlog"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
)

const (
	testPath = "/tmp/openstorage_driver_test"
)

var (
	testEnumerator volume.StoreEnumerator
	testLabels     = map[string]string{"Foo": "DEADBEEF"}
)

func TestCSISanity(t *testing.T) {

	kv, err := kvdb.New(mem.Name, "driver_test", []string{}, nil, dlog.Panicf)
	if err != nil {
		t.Fatalf("Failed to initialize KVDB")
	}
	if err := kvdb.SetInstance(kv); err != nil {
		t.Fatalf("Failed to set KVDB instance")
	}

	err = os.MkdirAll(testPath, 0744)
	if err != nil {
		t.Fatalf("Failed to create test path: %v", err)
	}

	if err := volumedrivers.Register("nfs", map[string]string{"path": testPath}); err != nil {
		t.Fatalf("Unable to start volume driver nfs")
	}

	// Initialize the cluster
	if err := cluster.Init(config.ClusterConfig{
		ClusterId:     "cluster",
		NodeId:        "node1",
		DefaultDriver: "nfs",
	}); err != nil {
		t.Fatalf("Unable to init cluster server: %v", err)
	}

	cm, err := cluster.Inst()
	if err != nil {
		t.Fatalf("Unable to find cluster instance: %v", err)
	}
	go func() {
		cm.Start(0, false)
	}()

	// Start CSI Server
	server, err := NewOsdCsiServer(&OsdCsiServerConfig{
		DriverName: "nfs",
		Net:        "tcp",
		Address:    "127.0.0.1:0",
		Cluster:    cm,
	})
	if err != nil {
		t.Fatalf("Unable to start csi server: %v", err)
	}
	server.Start()
	defer server.Stop()

	// Start CSI Sanity test
	sanity.Test(t, server.Address(), "/mnt/openstorage/mount/nfs")
}
