package types

import (
	"encoding/json"
	"testing"
)

func TestLegacyIDMarshalJSONAsString(t *testing.T) {
	var id LegacyID = 1234567890123
	b, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal LegacyID failed: %v", err)
	}

	if string(b) != `"1234567890123"` {
		t.Fatalf("unexpected marshal result: %s", string(b))
	}
}

func TestLegacyIDUnmarshalJSONString(t *testing.T) {
	var id LegacyID
	err := json.Unmarshal([]byte(`"42"`), &id)
	if err != nil {
		t.Fatalf("unmarshal string failed: %v", err)
	}

	if id != 42 {
		t.Fatalf("unexpected id: %d", id)
	}
}

func TestLegacyIDUnmarshalJSONNumber(t *testing.T) {
	var id LegacyID
	err := json.Unmarshal([]byte(`42`), &id)
	if err != nil {
		t.Fatalf("unmarshal number failed: %v", err)
	}

	if id != 42 {
		t.Fatalf("unexpected id: %d", id)
	}
}

func TestLegacyIDUnmarshalInvalid(t *testing.T) {
	var id LegacyID
	err := json.Unmarshal([]byte(`"abc"`), &id)
	if err == nil {
		t.Fatal("expected error for invalid legacy id")
	}
}
