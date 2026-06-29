package dqutil

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hardenedClientTrusting returns the REAL hardened download client (no auth,
// redirects refused, compression disabled) but trusting srv's self-signed cert,
// so download behavior is exercised against an httptest TLS server.
func hardenedClientTrusting(srv *httptest.Server) *http.Client {
	hc := HardenedDownloadClient()
	hc.Transport.(*http.Transport).TLSClientConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig
	return hc
}

func mustURL(s string) *url.URL { u, _ := url.Parse(s); return u }

func TestResolveSQL(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "q.sql")
	require.NoError(t, os.WriteFile(fp, []byte("SELECT 2"), 0o600))

	got, err := ResolveSQL([]string{"SELECT 1"}, "")
	require.NoError(t, err)
	assert.Equal(t, "SELECT 1", got)

	got, err = ResolveSQL(nil, fp)
	require.NoError(t, err)
	assert.Equal(t, "SELECT 2", got)

	_, err = ResolveSQL([]string{"SELECT 1"}, fp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not both")

	_, err = ResolveSQL(nil, "")
	require.Error(t, err)

	// A blank positional is treated as "no SQL given".
	_, err = ResolveSQL([]string{"   "}, "")
	require.Error(t, err)
}

func TestBuildSubmitBody_RequiresOutputTarget(t *testing.T) {
	b, err := BuildSubmitBody("SELECT 1", &SubmitFlags{OutputFormat: "JSON", Compression: "NONE"})
	require.NoError(t, err)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &body))
	assert.Equal(t, "SELECT 1", body["query"])
	out, ok := body["output"].(map[string]interface{})
	require.True(t, ok, "output object must be present")
	assert.Equal(t, "S3", out["target"])
}

func TestHardenedDownloadClient_Config(t *testing.T) {
	hc := HardenedDownloadClient()
	tr, ok := hc.Transport.(*http.Transport)
	require.True(t, ok)
	assert.True(t, tr.DisableCompression, "must disable compression to preserve exact bytes")
	require.NotNil(t, hc.CheckRedirect)
	err := hc.CheckRedirect(&http.Request{URL: mustURL("https://elsewhere/x")}, nil)
	assert.Error(t, err, "redirects must be refused")
}

func TestDownloadStream_NoAuth_PreservesBytes(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"), "download must not send Authorization off-host")
		w.Write([]byte("raw-result-bytes"))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	err := DownloadStream(context.Background(), srv.URL+"/r", &buf, hardenedClientTrusting(srv))
	require.NoError(t, err)
	assert.Equal(t, "raw-result-bytes", buf.String())
}

func TestDownloadStream_PreservesGzipBytes(t *testing.T) {
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte("hello"))
	require.NoError(t, zw.Close())
	gzBytes := gz.Bytes()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(gzBytes)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	err := DownloadStream(context.Background(), srv.URL, &buf, hardenedClientTrusting(srv))
	require.NoError(t, err)
	// DisableCompression means Go must NOT transparently decompress: the raw
	// gzip bytes are preserved (a --compression GZIP result stays compressed).
	assert.Equal(t, gzBytes, buf.Bytes())
}

func TestDownloadStream_RefusesRedirect(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect target must not be reached")
	}))
	defer target.Close()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer srv.Close()

	var buf bytes.Buffer
	err := DownloadStream(context.Background(), srv.URL, &buf, hardenedClientTrusting(srv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redirect")
}

func TestDownloadStream_RejectsNonHTTPS(t *testing.T) {
	var buf bytes.Buffer
	err := DownloadStream(context.Background(), "http://example.com/r", &buf, http.DefaultClient)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https")
}

func TestDownloadStream_RejectsNon2xx(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	var buf bytes.Buffer
	err := DownloadStream(context.Background(), srv.URL, &buf, hardenedClientTrusting(srv))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestDownloadToFile_LeavesExistingFileOnFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	dst := filepath.Join(t.TempDir(), "out.csv")
	require.NoError(t, os.WriteFile(dst, []byte("PRE-EXISTING"), 0o600))

	err := DownloadToFile(context.Background(), srv.URL, dst, nil, hardenedClientTrusting(srv))
	require.Error(t, err)
	b, rerr := os.ReadFile(dst)
	require.NoError(t, rerr)
	assert.Equal(t, "PRE-EXISTING", string(b), "a failed download must leave the existing file intact")
}

func TestDownloadToFile_Stdout(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("streamed"))
	}))
	defer srv.Close()
	var out bytes.Buffer
	err := DownloadToFile(context.Background(), srv.URL, "-", &out, hardenedClientTrusting(srv))
	require.NoError(t, err)
	assert.Equal(t, "streamed", out.String())
}

func TestDownloadToFile_Success(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("DATA"))
	}))
	defer srv.Close()
	dst := filepath.Join(t.TempDir(), "out.bin")
	require.NoError(t, DownloadToFile(context.Background(), srv.URL, dst, nil, hardenedClientTrusting(srv)))
	b, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, "DATA", string(b))
}

func TestBuildSubmitBody_AllOptions(t *testing.T) {
	b, err := BuildSubmitBody("SELECT 1", &SubmitFlags{
		OutputFormat: "DSV", Compression: "GZIP", ColumnSeparator: "|",
		Source: "WAREHOUSE", ReadDeleted: true, UseIndexJoin: true,
	})
	require.NoError(t, err)
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(b, &body))
	assert.Equal(t, "DSV", body["outputFormat"])
	assert.Equal(t, "GZIP", body["compression"])
	assert.Equal(t, "|", body["columnSeparator"])
	assert.Equal(t, "WAREHOUSE", body["sourceData"])
	assert.Equal(t, true, body["readDeleted"])
	assert.Equal(t, true, body["useIndexJoin"])
}

func TestUnwrapData(t *testing.T) {
	d := UnwrapData(map[string]interface{}{"data": map[string]interface{}{"id": "x"}})
	assert.Equal(t, "x", d["id"])
	// No data envelope → returns the raw map unchanged.
	raw := map[string]interface{}{"id": "y"}
	assert.Equal(t, raw, UnwrapData(raw))
}

func TestDecodeData(t *testing.T) {
	d, err := DecodeData([]byte(`{"data":{"id":"z"}}`))
	require.NoError(t, err)
	assert.Equal(t, "z", d["id"])
	_, err = DecodeData([]byte(`not json`))
	require.Error(t, err)
}

func TestDetailFields(t *testing.T) {
	got := map[string]string{}
	// outputRows/processingTime as JSON numbers (float64): must render as plain
	// decimals, NOT scientific notation.
	for _, fld := range DetailFields(map[string]interface{}{
		"id": "a", "queryStatus": "completed", "outputRows": float64(1000000), "processingTime": float64(2500), "dataFile": "u",
	}) {
		got[fld.Key] = fld.Value
	}
	assert.Equal(t, "a", got["ID"])
	assert.Equal(t, "completed", got["Status"])
	assert.Equal(t, "1000000", got["Output Rows"], "large counts must render as plain decimals, not scientific notation")
	assert.Equal(t, "2500", got["Processing Time"])
	assert.Equal(t, "u", got["Data File"])
}

func TestIsTerminalStatus(t *testing.T) {
	for _, s := range []string{"completed", "failed", "cancelled", "Canceled"} {
		assert.True(t, IsTerminalStatus(s), s)
	}
	for _, s := range []string{"accepted", "in_progress", ""} {
		assert.False(t, IsTerminalStatus(s), s)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	assert.Equal(t, "b", FirstNonEmpty("", "b", "c"))
	assert.Equal(t, "", FirstNonEmpty("", ""))
}

func TestJobPath(t *testing.T) {
	assert.Equal(t, "/query/jobs/a%2Fb", JobPath("a/b"))
}

func TestAddSubmitFlagsAndCompletions(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	var sf SubmitFlags
	AddSubmitFlags(cmd.Flags(), &sf)
	RegisterSubmitCompletions(cmd)
	for _, name := range []string{"file", "output-format", "compression", "column-separator", "source", "read-deleted", "use-index-join", "idempotency-key"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag %s must be registered", name)
	}
}
