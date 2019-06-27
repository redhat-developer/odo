package volume

import (
	"crypto/tls"
	"encoding/json"
	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vol *api.Volume

		json.NewEncoder(w).Encode(vol)
	}))

	defer ts.Close()

	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	clnt.SetTLS(&tls.Config{InsecureSkipVerify: true})

	_, err = VolumeDriver(clnt).Inspect([]string{"12345"})

	require.NoError(t, err)
}
