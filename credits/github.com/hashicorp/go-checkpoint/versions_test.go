package checkpoint

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestVersions(t *testing.T) {
	t.Skip("endpoint does not exist yet")

	expected := &VersionsResponse{
		Service:   "test.v1",
		Product:   "test",
		Minimum:   "1.0",
		Excluding: []string{"1.3"},
		Maximum:   "2.0",
	}

	actual, err := Versions(&VersionsParams{
		Service: "test.v1",
		Product: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestVersions_timeout(t *testing.T) {
	t.Skip("endpoint does not exist yet")

	os.Setenv("CHECKPOINT_TIMEOUT", "50")
	defer os.Setenv("CHECKPOINT_TIMEOUT", "")

	expected := "Client.Timeout exceeded while awaiting headers"

	_, err := Versions(&VersionsParams{
		Service: "test.v1",
		Product: "test",
	})

	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected a timeout error, got: %v", err)
	}
}

func TestVersions_disabled(t *testing.T) {
	t.Skip("endpoint does not exist yet")

	os.Setenv("CHECKPOINT_DISABLE", "1")
	defer os.Setenv("CHECKPOINT_DISABLE", "")

	expected := &CheckResponse{}

	actual, err := Versions(&VersionsParams{
		Service: "test.v1",
		Product: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}
