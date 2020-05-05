package fake

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
)

func TestHookCreateDelete(t *testing.T) {

	client, _ := NewDefault()

	in := &scm.HookInput{
		Target: "https://example.com",
		Name:   "test",
	}

	// create a hook
	createdHook, _, err := client.Repositories.CreateHook(context.Background(), "foo/repo", in)
	if err != nil {
		t.Fatal(err)
	}

	id := createdHook.ID
	if id == "" {
		t.Fatal("created hook must have an ID")
	}

	// list to verify created hook
	hooks, _, err := client.Repositories.ListHooks(context.Background(), "foo/repo", scm.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(hooks) != 1 {
		t.Fatal("expect one hook")
	}

	if diff := cmp.Diff(id, hooks[0].ID); diff != "" {
		t.Fatalf("hook id mismatch got\n%s", diff)
	}

	// delete by hook ID
	_, err = client.Repositories.DeleteHook(context.Background(), "foo/repo", id)
	if err != nil {
		t.Fatal(err)
	}

	// list to verify deletion
	hooks, _, err = client.Repositories.ListHooks(context.Background(), "foo/repo", scm.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}

	if len(hooks) != 0 {
		t.Fatal("expect no hooks")
	}
}
