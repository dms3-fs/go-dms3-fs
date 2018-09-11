package core_test

import (
	"testing"

	core "github.com/dms3-fs/go-dms3-fs/core"
	coremock "github.com/dms3-fs/go-dms3-fs/core/mock"
	path "github.com/dms3-fs/go-path"
)

func TestResolveNoComponents(t *testing.T) {
	n, err := coremock.NewMockNode()
	if n == nil || err != nil {
		t.Fatal("Should have constructed a mock node", err)
	}

	_, err = core.Resolve(n.Context(), n.Namesys, n.Resolver, path.Path("/dms3ns/"))
	if err != path.ErrNoComponents {
		t.Fatal("Should error with no components (/dms3ns/).", err)
	}

	_, err = core.Resolve(n.Context(), n.Namesys, n.Resolver, path.Path("/dms3fs/"))
	if err != path.ErrNoComponents {
		t.Fatal("Should error with no components (/dms3fs/).", err)
	}

	_, err = core.Resolve(n.Context(), n.Namesys, n.Resolver, path.Path("/../.."))
	if err != path.ErrBadPath {
		t.Fatal("Should error with invalid path.", err)
	}
}
