// Package graphql provides a GraphQL interface to ProbeChain node data.
// This is a minimal stub — GraphQL is not supported in Rydberg.
package graphql

import (
	"github.com/probechain/go-probe/internal/probeapi"
	"github.com/probechain/go-probe/node"
)

// New registers the GraphQL service against a node. Stub: returns nil (no-op).
func New(stack *node.Node, backend probeapi.Backend, cors, vhosts []string) error {
	return nil
}
