package helper

import (
	"hash/crc32"
	"slices"
	"sort"
)

// ConsistentHash defines the hash ring structure
type ConsistentHash struct {
	nodes   []uint32          // sorted hash ring
	nodeMap map[uint32]string // map from hash to node ID
}

// NewConsistentHash initializes the hash ring
func NewConsistentHash() *ConsistentHash {
	return &ConsistentHash{
		nodes:   []uint32{},
		nodeMap: make(map[uint32]string),
	}
}

// hashKey hashes the key using SHA256 and returns the first 4 bytes as uint32
// func hashKey(key string) uint32 {
// 	h := sha256.Sum256([]byte(key))
// 	return (uint32(h[0]) << 24) | (uint32(h[1]) << 16) | (uint32(h[2]) << 8) | uint32(h[3])
// }

// hashKey hashes the input using crc32
func hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}

// AddNode adds a node to the hash ring
func (c *ConsistentHash) AddNode(nodeID string) {
	hash := hashKey(nodeID)
	c.nodes = append(c.nodes, hash)
	c.nodeMap[hash] = nodeID
	slices.Sort(c.nodes)
}

// RemoveNode removes a node from the hash ring
func (c *ConsistentHash) RemoveNode(nodeID string) {
	hash := hashKey(nodeID)
	delete(c.nodeMap, hash)

	// Remove from sorted slice
	for i, h := range c.nodes {
		if h == hash {
			c.nodes = slices.Delete(c.nodes, i, i+1)
			break
		}
	}
}

// GetNode returns the node responsible for the given key
func (c *ConsistentHash) GetNode(key string) string {
	if len(c.nodes) == 0 {
		return ""
	}
	hash := hashKey(key)
	// Binary search for the first node >= hash
	idx := sort.Search(len(c.nodes), func(i int) bool { return c.nodes[i] >= hash })
	if idx == len(c.nodes) {
		idx = 0 // wrap around
	}
	return c.nodeMap[c.nodes[idx]]
}
