package verifycase

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/wyw14/cry-110/internal/api"
	"github.com/wyw14/cry-110/internal/model"
)

func TestRotorPinRejectsTorsionalRebound(t *testing.T) {
	start := time.Now().UTC()
	system, err := api.NewSystem(filepath.Join(t.TempDir(), "state"), start)
	if err != nil {
		t.Fatal(err)
	}
	server := api.NewServer(system)
	request := httptest.NewRequest(http.MethodPost, "/api/isolation", bytes.NewBufferString(`{"turbine_id":"WTG-110","reason":"main bearing inspection"}`))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("isolation status = %d, body = %s", response.Code, response.Body.String())
	}
	for index := 0; index < 5; index++ {
		at := start.Add(time.Duration(index) * 500 * time.Millisecond)
		err := system.Rotor.Record(model.RotorTelemetry{
			AverageRPM: 0.01, InstantRPM: 0.01, ShaftTorqueNm: 20,
			EncoderDegrees: 15, ObservedAt: at,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	reboundAt := start.Add(2100 * time.Millisecond)
	if err := system.Rotor.Record(model.RotorTelemetry{
		AverageRPM: 0.01, InstantRPM: 0.42, ShaftTorqueNm: 930,
		EncoderDegrees: 15.1, ObservedAt: reboundAt,
	}); err != nil {
		t.Fatal(err)
	}
	if err := system.Rotor.InsertPin(reboundAt); err == nil {
		t.Fatal("mechanical pin was allowed during a torsional rebound")
	}
}
