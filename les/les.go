// Package les implements the Light ProbeChain Subprotocol.
// This is a minimal stub — light client mode is not supported in Rydberg.
package les

import (
	"errors"

	"github.com/probechain/go-probe/internal/probeapi"
	"github.com/probechain/go-probe/node"
	"github.com/probechain/go-probe/probe"
	"github.com/probechain/go-probe/probe/probeconfig"
)

// LightProbe implements the light client.
type LightProbe struct {
	ApiBackend probeapi.Backend
}

// New creates a light client instance. Stub: returns error (not supported).
func New(stack *node.Node, cfg *probeconfig.Config) (*LightProbe, error) {
	return nil, errors.New("light client mode is not supported in Rydberg testnet")
}

// LesServer implements the LES server.
type LesServer struct{}

// NewLesServer creates a LES server. Stub: returns error (not supported).
func NewLesServer(stack *node.Node, backend *probe.Probeum, cfg *probeconfig.Config) (*LesServer, error) {
	return nil, errors.New("LES server is not supported in Rydberg testnet")
}
