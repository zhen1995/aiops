package knowledge

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aiops/models"
)

func testEmb() *models.LLMConfig {
	return &models.LLMConfig{BaseURL: "http://emb/v1", APIKey: "k", Model: "m"}
}

func TestIndex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/knowledge/index" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("document_id") != "doc-1" || r.FormValue("model") != "m" {
			t.Errorf("missing form fields: %v", r.Form)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"document_id":"doc-1","chunks":3}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	n, err := c.Index(context.Background(), "doc-1", "标题", testEmb(), "a.md", strings.NewReader("# hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("chunks = %d, want 3", n)
	}
}

func TestIndexUnsupportedType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(`{"detail":"不支持的文件类型: exe"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Index(context.Background(), "d", "t", testEmb(), "a.exe", strings.NewReader("MZ"))
	if err == nil || !strings.Contains(err.Error(), "不支持的文件类型") {
		t.Errorf("err = %v, want 包含 不支持的文件类型", err)
	}
}

func TestRetrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[{"chunk_id":"doc-1-0","document_id":"doc-1","chunk_index":0,"content":"MySQL 排查","title":"手册","score":0.91}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	hits, err := c.Retrieve(context.Background(), "连接超时", 5, testEmb())
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].DocumentID != "doc-1" || hits[0].Score != 0.91 {
		t.Errorf("unexpected hits: %+v", hits)
	}
}

func TestDeleteDocument(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/knowledge/documents/doc-1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Write([]byte(`{"deleted":true}`))
	}))
	defer srv.Close()

	if err := NewClient(srv.URL).DeleteDocument(context.Background(), "doc-1"); err != nil {
		t.Fatal(err)
	}
}

var _ = io.Discard
