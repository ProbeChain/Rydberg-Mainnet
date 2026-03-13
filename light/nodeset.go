package light

import (
	"errors"

	"github.com/probechain/go-probe/probedb"
	"github.com/probechain/go-probe/rlp"
)

// NodeList stores an ordered list of trie nodes for Merkle proofs.
type NodeList []rlp.RawValue

// NodeSet converts the node list to a NodeSet.
func (n NodeList) NodeSet() *NodeSet {
	db := NewNodeSet()
	for i, node := range n {
		key := make([]byte, 8)
		key[0] = byte(i >> 24)
		key[1] = byte(i >> 16)
		key[2] = byte(i >> 8)
		key[3] = byte(i)
		db.Put(key, []byte(node))
	}
	return db
}

// NodeSet stores a set of trie nodes for Merkle proofs.
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
	k := string(key)
	if _, ok := db.nodes[k]; ok {
		return nil
	}
	db.nodes[k] = value
	db.order = append(db.order, k)
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
	return nil, errors.New("not found")
}

// Has returns whether a node is in the set.
func (db *NodeSet) Has(key []byte) (bool, error) {
	_, ok := db.nodes[string(key)]
	return ok, nil
}

// KeyCount returns the number of nodes in the set.
func (db *NodeSet) KeyCount() int {
	return len(db.nodes)
}

// NodeList returns the list of trie nodes.
func (db *NodeSet) NodeList() NodeList {
	var list NodeList
	for _, k := range db.order {
		list = append(list, db.nodes[k])
	}
	return list
}

// Store writes the contents of the set to the given database.
func (db *NodeSet) Store(target probedb.KeyValueWriter) {
	for _, k := range db.order {
		target.Put([]byte(k), db.nodes[k])
	}
}
