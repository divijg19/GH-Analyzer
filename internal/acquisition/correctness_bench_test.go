package acquisition

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkFetchRepos measures acquisition of a multi-page repository list
// (per_page=100, three pages). Represents the corrected pagination path.
func BenchmarkFetchRepos(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", `<http://`+r.Host+`/users/u/repos?page=2>; rel="next"`)
			writeRepos(w, 100)
		case "2":
			w.Header().Set("Link", `<http://`+r.Host+`/users/u/repos?page=3>; rel="next"`)
			writeRepos(w, 100)
		default:
			writeRepos(w, 100)
		}
	}))
	defer srv.Close()

	client := testClient(srv.URL)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.FetchRepos(context.Background(), "u"); err != nil {
			b.Fatal(err)
		}
	}
}
