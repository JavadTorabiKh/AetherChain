package network

import (
	"fmt"
	"net"
	"sync"
	"time"

	"javadtorabikh/Aetherchain/config"
	"javadtorabikh/Aetherchain/blockchain"
)

// Node represents a network node in the AetherChain network
type Node struct {
	config     *config.Config
	blockchain *blockchain.Blockchain
	
	// Network properties
	listener   net.Listener
	peers      map[string]*Peer
	peerMutex  sync.RWMutex
	
	// Node state
	isRunning  bool
	stopCh     chan struct{}
}

// Peer represents a connected peer node
type Peer struct {
	ID        string
	Address   string
	Conn      net.Conn
	Connected bool
	LastSeen  time.Time
}

// NewNode creates a new network node
func NewNode(cfg *config.Config, bc *blockchain.Blockchain) *Node {
	return &Node{
		config:     cfg,
		blockchain: bc,
		peers:      make(map[string]*Peer),
		stopCh:     make(chan struct{}),
	}
}

// Start begins listening for incoming connections
func (n *Node) Start() error {
	address := fmt.Sprintf("%s:%d", n.config.Host, n.config.Port)
	
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start node: %v", err)
	}
	
	n.listener = listener
	n.isRunning = true
	
	fmt.Printf("🔌 Node listening on %s\n", address)
	
	// Start accepting connections
	go n.acceptConnections()
	
	// Connect to bootstrap nodes
	go n.connectToBootstrapNodes()
	
	// Start peer maintenance
	go n.peerMaintenance()
	
	return nil
}

// Stop gracefully shuts down the node
func (n *Node) Stop() {
	n.isRunning = false
	close(n.stopCh)
	
	if n.listener != nil {
		n.listener.Close()
	}
	
	// Close all peer connections
	n.peerMutex.Lock()
	for _, peer := range n.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}
	n.peers = make(map[string]*Peer)
	n.peerMutex.Unlock()
	
	fmt.Println("🔌 Node stopped")
}

// acceptConnections handles incoming connections
func (n *Node) acceptConnections() {
	for n.isRunning {
		conn, err := n.listener.Accept()
		if err != nil {
			if n.isRunning {
				fmt.Printf("Error accepting connection: %v\n", err)
			}
			continue
		}
		
		go n.handleConnection(conn)
	}
}

// handleConnection processes a new connection
func (n *Node) handleConnection(conn net.Conn) {
	peerAddress := conn.RemoteAddr().String()
	fmt.Printf("🔗 New connection from %s\n", peerAddress)
	
	peer := &Peer{
		ID:        generatePeerID(),
		Address:   peerAddress,
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
	}
	
	n.addPeer(peer)
	
	// Handle peer communication
	n.handlePeerCommunication(peer)
}

// handlePeerCommunication manages communication with a peer
func (n *Node) handlePeerCommunication(peer *Peer) {
	defer func() {
		peer.Connected = false
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		n.removePeer(peer.ID)
		fmt.Printf("🔌 Disconnected from peer %s\n", peer.Address)
	}()
	
	buffer := make([]byte, 4096)
	
	for n.isRunning && peer.Connected {
		// Set read timeout
		peer.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		
		n, err := peer.Conn.Read(buffer)
		if err != nil {
			if n.isRunning {
				fmt.Printf("Error reading from peer %s: %v\n", peer.Address, err)
			}
			return
		}
		
		if n > 0 {
			peer.LastSeen = time.Now()
			n.handleMessage(peer, buffer[:n])
		}
	}
}

// handleMessage processes incoming messages from peers
func (n *Node) handleMessage(peer *Peer, data []byte) {
	// Parse and handle different message types
	// This is a simplified implementation
	fmt.Printf("📨 Received message from %s: %s\n", peer.Address, string(data))
	
	// Echo back for now
	response := fmt.Sprintf("Echo: %s", string(data))
	peer.Conn.Write([]byte(response))
}

// connectToBootstrapNodes connects to bootstrap nodes
func (n *Node) connectToBootstrapNodes() {
	for _, bootstrapNode := range n.config.BootstrapNodes {
		go n.connectToNode(bootstrapNode)
	}
}

// connectToNode attempts to connect to a specific node
func (n *Node) connectToNode(address string) {
	conn, err := net.DialTimeout("tcp", address, n.config.PeerTimeout)
	if err != nil {
		fmt.Printf("Failed to connect to bootstrap node %s: %v\n", address, err)
		return
	}
	
	fmt.Printf("🔗 Connected to bootstrap node %s\n", address)
	
	peer := &Peer{
		ID:        generatePeerID(),
		Address:   address,
		Conn:      conn,
		Connected: true,
		LastSeen:  time.Now(),
	}
	
	n.addPeer(peer)
	go n.handlePeerCommunication(peer)
}

// peerMaintenance performs maintenance tasks on peers
func (n *Node) peerMaintenance() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			n.cleanupDeadPeers()
		case <-n.stopCh:
			return
		}
	}
}

// cleanupDeadPeers removes peers that haven't been seen recently
func (n *Node) cleanupDeadPeers() {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	
	for id, peer := range n.peers {
		if time.Since(peer.LastSeen) > n.config.PeerTimeout {
			if peer.Conn != nil {
				peer.Conn.Close()
			}
			delete(n.peers, id)
			fmt.Printf("🧹 Removed dead peer: %s\n", peer.Address)
		}
	}
}

// addPeer adds a peer to the peer list
func (n *Node) addPeer(peer *Peer) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	
	n.peers[peer.ID] = peer
	fmt.Printf("👥 Added peer: %s (Total: %d)\n", peer.Address, len(n.peers))
}

// removePeer removes a peer from the peer list
func (n *Node) removePeer(peerID string) {
	n.peerMutex.Lock()
	defer n.peerMutex.Unlock()
	
	delete(n.peers, peerID)
}

// GetPeerCount returns the number of connected peers
func (n *Node) GetPeerCount() int {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	
	return len(n.peers)
}

// BroadcastMessage sends a message to all connected peers
func (n *Node) BroadcastMessage(message []byte) {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()
	
	for _, peer := range n.peers {
		if peer.Connected {
			_, err := peer.Conn.Write(message)
			if err != nil {
				fmt.Printf("Failed to send message to peer %s: %v\n", peer.Address, err)
			}
		}
	}
}

func generatePeerID() string {
	return fmt.Sprintf("peer_%d", time.Now().UnixNano())
}