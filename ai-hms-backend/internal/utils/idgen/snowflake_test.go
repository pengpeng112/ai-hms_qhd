package idgen

import "testing"

func TestNextIDUnique(t *testing.T) {
	g, err := NewGenerator(1)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ids := make(map[int64]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := g.NextID()
		if _, exists := ids[id]; exists {
			t.Fatalf("duplicate ID generated: %d", id)
		}
		ids[id] = struct{}{}
	}
}

func TestParseNodeIDFromEnvInvalidFallback(t *testing.T) {
	t.Setenv("SNOWFLAKE_NODE_ID", "invalid")
	if got := parseNodeIDFromEnv(); got != defaultNodeID {
		t.Fatalf("parseNodeIDFromEnv() = %d, want %d", got, defaultNodeID)
	}
}

func TestParseNodeIDFromEnvOutOfRangeFallback(t *testing.T) {
	t.Setenv("SNOWFLAKE_NODE_ID", "2048")
	if got := parseNodeIDFromEnv(); got != defaultNodeID {
		t.Fatalf("parseNodeIDFromEnv() = %d, want %d", got, defaultNodeID)
	}
}
