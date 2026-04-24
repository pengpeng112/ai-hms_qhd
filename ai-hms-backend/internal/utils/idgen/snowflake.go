package idgen

import (
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

const (
	defaultNodeID int64 = 1
	maxNodeID     int64 = 1023
)

type Generator struct {
	node *snowflake.Node
}

func NewGenerator(nodeID int64) (*Generator, error) {
	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}

	return &Generator{node: node}, nil
}

func (g *Generator) NextID() int64 {
	return g.node.Generate().Int64()
}

var (
	defaultGenerator     *Generator
	defaultGeneratorOnce sync.Once
	defaultGeneratorErr  error
)

// NextID returns a snowflake int64 ID using a process-wide singleton generator.
func NextID() (int64, error) {
	g, err := getDefaultGenerator()
	if err != nil {
		return 0, err
	}

	return g.NextID(), nil
}

func getDefaultGenerator() (*Generator, error) {
	defaultGeneratorOnce.Do(func() {
		nodeID := parseNodeIDFromEnv()
		defaultGenerator, defaultGeneratorErr = NewGenerator(nodeID)
	})

	return defaultGenerator, defaultGeneratorErr
}

func parseNodeIDFromEnv() int64 {
	raw := os.Getenv("SNOWFLAKE_NODE_ID")
	if raw == "" {
		return defaultNodeID
	}

	nodeID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || nodeID < 0 || nodeID > maxNodeID {
		return defaultNodeID
	}

	return nodeID
}
