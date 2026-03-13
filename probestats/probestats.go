// Package probestats implements the network stats reporting service.
// This is a minimal stub — probestats is not supported in Rydberg.
package probestats

import (
	"github.com/probechain/go-probe/consensus"
	"github.com/probechain/go-probe/internal/probeapi"
	"github.com/probechain/go-probe/node"
)

// New creates a stats reporting daemon. Stub: returns nil (no-op).
func New(stack *node.Node, backend probeapi.Backend, engine consensus.Engine, url string) error {
	return nil
}
