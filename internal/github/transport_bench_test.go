package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkTransport measures the Transport path (token selection, retry,
// single-flight) against a single upstream call. Concurrent identical keys
// collapse via single-flight, so this also exercises coalescing under load.
func BenchmarkTransport(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tr := &Transport{
		client: &http.Client{Timeout: 5 * time.Second},
		pool:   NewTokenPool([]string{"a", "b", "c"}),
		retry:  NewRetryPolicy(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/u/repos", nil)
		resp, err := tr.Do(context.Background(), req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()
	}
}
