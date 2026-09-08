package HttpService

import (
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsJSON"
	"net/http"
	"net/http/httptest"
	"testing"
)

type smokeCallback string

func (c smokeCallback) Process(http.ResponseWriter, *http.Request, *MarsJSON.JSONObject, []string, *MarsJSON.JSONObject, string) []byte {
	return []byte(c)
}
func TestSmokeRouteLifecycle(t *testing.T) {
	s := Create(0, 0, "", "", "")
	for _, body := range []string{"first", "replacement"} {
		s.AddRestfulAPI("smoke", smokeCallback(body))
		w := httptest.NewRecorder()
		s._Mux.ServeHTTP(w, httptest.NewRequest("GET", "/smoke/", nil))
		if w.Code != 200 || w.Body.String() != body {
			t.Fatalf("response: %d %q", w.Code, w.Body.String())
		}
		s.RemoveRestfulAPI("smoke")
		w = httptest.NewRecorder()
		s._Mux.ServeHTTP(w, httptest.NewRequest("GET", "/smoke/", nil))
		if w.Code != 404 {
			t.Fatalf("removed route: %d", w.Code)
		}
	}
}

func TestSmokeHTTPListener(t *testing.T) {
	s := Create(0, 0, "", "", "")
	server := httptest.NewServer(s._Mux)
	defer server.Close()
	s.AddRestfulAPI("smoke", smokeCallback("live"))
	resp, err := server.Client().Get(server.URL + "/smoke/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HTTP: %d", resp.StatusCode)
	}
}
func TestSmokeStartupRollback(t *testing.T) {
	s := Create(1, 2, "missing-certificate.pem", "", "")
	s._HttpServer.Addr = "127.0.0.1:0"
	s._HttpsServer.Addr = "127.0.0.1:0"
	if err := s.RunWithError(); err == nil {
		t.Fatal("無效 TLS 憑證應啟動失敗")
	}
	if !s.Close() {
		t.Fatal("未啟動的服務應可安全關閉")
	}
}
