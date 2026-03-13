// Package light implements on-demand retrieval capable state and chain objects
// for the ProbeChain Light Probe Subprotocol.
// This is a minimal stub providing only types needed by snap sync.
package light

import (
	"errors"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/probedb"
	"github.com/probechain/go-probe/rlp"
	"golang.org/x/crypto/sha3"
)

var errNotFound = errors.New("not found")

// NodeSet stores a set of trie nodes. It implements trie.DatabaseReader.
type NodeSet struct {
	nodes map[string][]byte
	order []string
}

// NewNodeSet creates an empty node set.
func NewNodeSet() *NodeSet {
	return &NodeSet{nodes: make(map[string][]byte)}
}

// Put stores a new node in the set.
func (db *NodeSet) Put(key []byte, value []byte) error {
	if _, ok := db.nodes[string(key)]; ok {
		return nil
	}
	keystr := string(key)
	db.nodes[keystr] = common.CopyBytes(value)
	db.order = append(db.order, keystr)
	return nil
}

// Delete removes a node from the set.
func (db *NodeSet) Delete(key []byte) error {
	delete(db.nodes, string(key))
	return nil
}

// Get returns a stored node.
func (db *NodeSet) Get(key []byte) ([]byte, error) {
	if entry, ok := db.nodes[string(key)]; ok {
		return entry, nil
	}
	return nil, errNotFound
}

// Has returns true if the node set contains the given key.
func (db *NodeSet) Has(key []byte) (bool, error) {
	_, ok := db.nodes[string(key)]
	return ok, nil
}

// KeyCount returns the number of nodes in the set.
func (db *NodeSet) KeyCount() int {
	return len(db.nodes)
}

// NodeList converts the node set to a NodeList.
func (db *NodeSet) NodeList() NodeList {
	result := make(NodeList, 0, len(db.order))
	for _, key := range db.order {
		result = append(result, db.nodes[key])
	}
	return result
}

// NodeList stores an ordered list of trie nodes (RLP-encoded blobs).
type NodeList []rlp.RawValue

// NodeSet converts the node list to a NodeSet (key = keccak256(blob)).
func (n NodeList) NodeSet() *NodeSet {
	db := NewNodeSet()
	for _, node := range n {
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(node)
		var key common.Hash
		hasher.Sum(key[:0])
		db.nodes[string(key[:])] = node
		db.order = append(db.order, string(key[:]))
	}
	return db
}

// Store adds the contents of the node list to an existing database.
func (n NodeList) Store(db probedb.KeyValueWriter) {
	for _, node := range n {
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(node)
		var key common.Hash
		hasher.Sum(key[:0])
		db.Put(key[:], node)
	}
}
