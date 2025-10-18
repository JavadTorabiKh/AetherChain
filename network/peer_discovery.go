package network

import (
	"fmt"
	"time"
)

// PeerDiscovery manages peer discovery and network connectivity
type PeerDiscovery struct {
	node    *Node
	peers   map[string]time.Time // peer address -> last discovered
}

// NewPeerDiscovery creates a new peer discovery instance
func NewPeerDiscovery(node *Node) *PeerDiscovery {
	return &PeerDiscovery{
		node:  node,
		peers: make(map[string]time.Time),
	}
}

// DiscoverPeers attempts to discover new peers in the network
func (pd *PeerDiscovery) DiscoverPeers() {
	fmt.Println("🔍 Discovering peers...")
	
	// In a real implementation, this would:
	// 1. Query known peers for their peer lists
	// 2. Use DNS-based peer discovery
	// 3. Use hardcoded bootstrap nodes
	// 4. Use peer exchange protocols
	
	// For now, we'll just print a message
	fmt.Println("Peer discovery completed")
}

// AddDiscoveredPeer adds a discovered peer to the discovery list
func (pd *PeerDiscovery) AddDiscoveredPeer(address string) {
	pd.peers[address] = time.Now()
	fmt.Printf("📝 Discovered new peer: %s\n", address)
}

// GetDiscoveredPeers returns all discovered peers
func (pd *PeerDiscovery) GetDiscoveredPeers() []string {
	var peers []string
	for peer := range pd.peers {
		peers = append(peers, peer)
	}
	return peers
}

// CleanupOldPeers removes peers that were discovered too long ago
func (pd *PeerDiscovery) CleanupOldPeers(maxAge time.Duration) {
	now := time.Now()
	for peer, discovered := range pd.peers {
		if now.Sub(discovered) > maxAge {
			delete(pd.peers, peer)
			fmt.Printf("🧹 Removed old discovered peer: %s\n", peer)
		}
	}
}