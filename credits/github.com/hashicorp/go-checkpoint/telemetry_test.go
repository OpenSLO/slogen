package checkpoint

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReport_sendsRequest(t *testing.T) {
	expected := &ReportParams{
		Signature: "sig",
		Product:   "prod",
	}

	req, err := ReportRequest(expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer req.Body.Close()

	if !strings.HasSuffix(req.URL.Path, "/telemetry/prod") {
		t.Fatalf("expected url to include the product, got %s", req.URL.String())
	}

	var actual ReportParams
	if err := json.NewDecoder(req.Body).Decode(&actual); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if actual.Signature != expected.Signature {
		t.Fatalf("expected %#v, got %#v", expected, actual)
	}
}
