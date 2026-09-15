package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindAndDownload(t *testing.T) {
	asset := []byte("bundle content")
	sum := sha256.Sum256(asset)
	sums := fmt.Sprintf("%s  bundle.tar.gz\n", hex.EncodeToString(sum[:]))

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/o/r/releases/latest", "/repos/o/r/releases/tags/v1.0.0":
			if r.Header.Get("Authorization") != "Bearer tok" {
				t.Errorf("API request without token: %q", r.Header.Get("Authorization"))
			}
			fmt.Fprintf(w, `{"tag_name":"v1.0.0","assets":[
				{"name":"bundle.tar.gz","browser_download_url":"%[1]s/dl/bundle.tar.gz"},
				{"name":"SHA256SUMS","browser_download_url":"%[1]s/dl/SHA256SUMS"},
				{"name":"bad.tar.gz","browser_download_url":"%[1]s/dl/bundle.tar.gz"}]}`, srv.URL)
		case "/dl/bundle.tar.gz":
			if r.Header.Get("Authorization") != "" {
				t.Error("download request contains the token")
			}
			w.Write(asset)
		case "/dl/SHA256SUMS":
			fmt.Fprint(w, sums+strings.Repeat("0", 64)+"  bad.tar.gz\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := &Client{APIURL: srv.URL, Repo: "o/r", Token: "tok"}
	ctx := context.Background()

	for _, tag := range []string{"", "v1.0.0"} {
		rel, err := c.Find(ctx, tag)
		if err != nil || rel.Tag != "v1.0.0" || len(rel.Assets) != 3 {
			t.Fatalf("Find(%q) = %+v, %v", tag, rel, err)
		}
	}
	if _, err := c.Find(ctx, "v9.9.9"); err == nil || !strings.Contains(err.Error(), "no release with the tag v9.9.9") {
		t.Errorf("Find(v9.9.9) error = %v", err)
	}

	rel, _ := c.Find(ctx, "")
	data, err := c.DownloadVerified(ctx, rel, "bundle.tar.gz", 1024)
	if err != nil || string(data) != string(asset) {
		t.Errorf("DownloadVerified = %q, %v", data, err)
	}
	if _, err := c.DownloadVerified(ctx, rel, "bad.tar.gz", 1024); err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Errorf("DownloadVerified(bad) error = %v", err)
	}
	if _, err := c.Download(ctx, rel, "bundle.tar.gz", 3); err == nil {
		t.Error("Download over the limit: no error")
	}
	if _, err := c.Download(ctx, rel, "missing", 1024); err == nil {
		t.Error("Download of a missing asset: no error")
	}
}

func TestParseSums(t *testing.T) {
	good := strings.Repeat("a", 64) + "  file one\n" + strings.Repeat("B", 64) + " *bin\n\n"
	sums, err := ParseSums([]byte(good))
	if err != nil || sums["file one"] != strings.Repeat("a", 64) || sums["bin"] != strings.Repeat("b", 64) {
		t.Errorf("ParseSums = %v, %v", sums, err)
	}
	for _, bad := range []string{"abc  file", strings.Repeat("z", 64) + "  file", strings.Repeat("a", 64)} {
		if _, err := ParseSums([]byte(bad)); err == nil {
			t.Errorf("ParseSums(%q): no error", bad)
		}
	}
}
