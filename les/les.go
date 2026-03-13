package les

import (
	"fmt"

	"github.com/probechain/go-probe/node"
	"github.com/probechain/go-probe/probe"
	"github.com/probechain/go-probe/probe/probeconfig"
	"github.com/probechain/go-probe/probe/tracers"
)

// LightProbe implements the light client.
type LightProbe struct {
	ApiBackend tracers.Backend
}

// New creates a new light client (stub — not supported in Rydberg).
func New(stack *node.Node, cfg *probeconfig.Config) (*LightProbe, error) {
	return nil, fmt.Errorf("light sync mode is not supported in Rydberg mainnet")
}

// LesServer implements the LES server.
type LesServer struct{}

// NewLesServer creates a new LES server (stub — not supported in Rydberg).
func NewLesServer(stack *node.Node, backend *probe.Probeum, cfg *probeconfig.Config) (*LesServer, error) {
	return nil, fmt.Errorf("LES server is not supported in Rydberg mainnet")
}
