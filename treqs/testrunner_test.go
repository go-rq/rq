package treqs_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-rq/rq"
	"github.com/go-rq/rq/treqs"
)

func TestTreqs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"foo": "bar"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	treqs.RunDir(t, rq.WithEnvironment(context.Background(), map[string]string{
		"host": srv.URL,
	}), "../testdata", treqs.WithVerboseLogging)
}
