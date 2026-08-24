package verifycase

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-110/internal/api"
)

func TestIcingZonePersistsThroughControlledShedCycle(t *testing.T) {
	system, err := api.NewSystem(filepath.Join(t.TempDir(), "state"), time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	server := api.NewServer(system)
	request := httptest.NewRequest(http.MethodPost, "/api/incidents", bytes.NewBufferString(`{"kind":"icing","message":"ice accretion detected"}`))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("incident status = %d, body = %s", response.Code, response.Body.String())
	}
	if err := system.Icing.OnStopped(); err != nil {
		t.Fatal(err)
	}
	if err := system.Access.Enter(); err == nil {
		t.Fatal("tower-base access opened before the controlled ice-shed cycle")
	}
	if err := system.Icing.StartShedCycle(); err != nil {
		t.Fatal(err)
	}
	if err := system.Access.Enter(); err == nil {
		t.Fatal("tower-base access opened while ice could still shed")
	}
}
