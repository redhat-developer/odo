/*
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

package sanity

import (
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/api/client"
	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/cluster"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster [Cluster Tests]", func() {
	var (
		restClient *client.Client
		manager    cluster.Cluster
	)

	BeforeEach(func() {
		var err error
		restClient, err = clusterclient.NewClusterClient(osdAddress, cluster.APIVersion)
		Expect(err).ToNot(HaveOccurred())
		manager = clusterclient.ClusterManager(restClient)
	})

	AfterEach(func() {
	})

	Describe("Cluster Status", func() {
		It("should have OK status for the cluster", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking status the cluster")
			Expect(cluster.Id).NotTo(BeEmpty())
			Expect(cluster.NodeId).NotTo(BeEmpty())
			Expect(cluster.Nodes).NotTo(BeEmpty())
			Expect(cluster.Status).To(Equal(api.Status_STATUS_OK))
		})

		It("should have OK status for all nodes", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking status for each node")
			for _, n := range cluster.Nodes {
				Expect(n.Id).NotTo(BeEmpty())
				Expect(n.Hostname).NotTo(BeEmpty())
				Expect(n.Status).To(Equal(api.Status_STATUS_OK))
				Expect(n.Cpu).To(BeNumerically(">", 0))
				Expect(n.MemTotal).To(BeNumerically(">", 0))
				Expect(n.MemUsed).To(BeNumerically(">", 0))
				Expect(n.MemUsed).To(BeNumerically(">", 0))
			}
		})
	})

	Describe("Cluster Inspect", func() {
		It("should have ok inspecting each node", func() {
			By("Enumerating the nodes")
			cluster, err := manager.Enumerate()
			Expect(err).NotTo(HaveOccurred())

			By("checking inspecting node")
			for _, n := range cluster.Nodes {
				node, err := manager.Inspect(n.Id)
				Expect(err).NotTo(HaveOccurred())
				Expect(node.Id).To(Equal(n.Id))
				Expect(node.Status).To(Equal(n.Status))
				Expect(node.MgmtIp).To(Equal(n.MgmtIp))
				Expect(node.DataIp).To(Equal(n.DataIp))
				Expect(node.Hostname).To(Equal(n.Hostname))
			}
		})
	})
})
