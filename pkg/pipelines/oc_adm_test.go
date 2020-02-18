package pipelines

import (
	"testing"

	"k8s.io/apiserver/pkg/authentication/serviceaccount"

	v1 "github.com/openshift/api/security/v1"
	"github.com/openshift/client-go/security/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAdmPolicyAdd(t *testing.T) {
	var sccName = "privileged"

	sccObj := &v1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{Name: sccName},
		Users:      []string{serviceaccount.MakeUsername("ns1", "name1"), serviceaccount.MakeUsername("ns1", "name2")},
	}
	s := newFakeSCC([]runtime.Object{sccObj})
	retrievedSCCObj, err := s.client.Get(sccName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get scc %w", err)
	}
	if len(retrievedSCCObj.Users) != 2 {
		t.Fatalf("expected 2 users but got %d", len(retrievedSCCObj.Users))
	}

	s.addSCCToUser(sccName, "ns", "user2")
	retrievedSCCObj2, err2 := s.client.Get(sccName, metav1.GetOptions{})
	if err2 != nil {
		t.Fatalf("failed to get scc %w", err2)
	}
	if len(retrievedSCCObj2.Users) != 3 {
		t.Fatalf("expected 3 users but got %d", len(retrievedSCCObj2.Users))
	}
	var found = false
	for _, elem := range retrievedSCCObj2.Users {
		if serviceaccount.MakeUsername("ns", "user2") == elem {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("%s not found in SCC", serviceaccount.MakeUsername("ns", "user2"))
	}
}

// newFakeSCC create a Fake SCC
func newFakeSCC(objs []runtime.Object) *scc {
	// obtain fake client
	return &scc{
		client: fake.NewSimpleClientset(objs...).SecurityV1().SecurityContextConstraints(),
	}
}
