package network

import (
	"encoding/json"
	"fmt"
	"time"

	"Aetherchain/blockchain"
)

// MessageType represents different types of network messages
type MessageType string

const (
	MessageTypePing      MessageType = "ping"
	MessageTypePong      MessageType = "pong"
	MessageTypeGetBlocks MessageType = "get_blocks"
	MessageTypeBlocks    MessageType = "blocks"
	MessageTypeNewBlock  MessageType = "new_block"
	MessageTypeNewTx     MessageType = "new_tx"
	MessageTypeGetPeers  MessageType = "get_peers"
	MessageTypePeers     MessageType = "peers"
)

// NetworkMessage represents a message sent between nodes
type NetworkMessage struct {
	Type      MessageType     `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64          `json:"timestamp"`
	NodeID    string         `json:"node_id"`
	Version   string         `json:"version"`
}

// PingMessage data for ping messages
type PingMessage struct {
	Height    int    `json:"height"`
	BestHash  string `json:"best_hash"`
}

// PongMessage data for pong messages
type PongMessage struct {
	Height    int    `json:"height"`
	BestHash  string `json:"best_hash"`
}

// BlocksMessage data for sending blocks
type BlocksMessage struct {
	Blocks []*blockchain.Block `json:"blocks"`
}

// NewBlockMessage data for announcing new blocks
type NewBlockMessage struct {
	Block *blockchain.Block `json:"block"`
}

// NewTxMessage data for announcing new transactions
type NewTxMessage struct {
	Transaction *blockchain.Transaction `json:"transaction"`
}

// PeersMessage data for exchanging peer information
type PeersMessage struct {
	Peers []string `json:"peers"`
}

// MessageHandler handles incoming network messages
type MessageHandler struct {
	node *Node
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(node *Node) *MessageHandler {
	return &MessageHandler{
		node: node,
	}
}

// HandleMessage processes an incoming network message
func (mh *MessageHandler) HandleMessage(peer *Peer, rawData []byte) {
	var message NetworkMessage
	err := json.Unmarshal(rawData, &message)
	if err != nil {
		fmt.Printf("❌ Failed to parse message from %s: %v\n", peer.Address, err)
		return
	}

	fmt.Printf("📨 Received %s message from %s\n", message.Type, peer.Address)

	switch message.Type {
	case MessageTypePing:
		mh.handlePing(peer, message)
	case MessageTypePong:
		mh.handlePong(peer, message)
	case MessageTypeGetBlocks:
		mh.handleGetBlocks(peer, message)
	case MessageTypeBlocks:
		mh.handleBlocks(peer, message)
	case MessageTypeNewBlock:
		mh.handleNewBlock(peer, message)
	case MessageTypeNewTx:
		mh.handleNewTx(peer, message)
	case MessageTypeGetPeers:
		mh.handleGetPeers(peer, message)
	case MessageTypePeers:
		mh.handlePeers(peer, message)
	default:
		fmt.Printf("❌ Unknown message type: %s\n", message.Type)
	}
}

// handlePing processes ping messages
func (mh *MessageHandler) handlePing(peer *Peer, message NetworkMessage) {
	var pingData PingMessage
	if err := json.Unmarshal(message.Data, &pingData); err != nil {
		fmt.Printf("❌ Invalid ping data: %v\n", err)
		return
	}

	// Update peer information
	peer.LastSeen = time.Now()

	// Send pong response
	pongData := PongMessage{
		Height:   len(mh.node.blockchain.Chain),
		BestHash: mh.node.blockchain.GetLastBlock().Hash,
	}

	mh.sendMessage(peer, MessageTypePong, pongData)
}

// handlePong processes pong messages
func (mh *MessageHandler) handlePong(peer *Peer, message NetworkMessage) {
	var pongData PongMessage
	if err := json.Unmarshal(message.Data, &pongData); err != nil {
		fmt.Printf("❌ Invalid pong data: %v\n", err)
		return
	}

	// Update peer information
	peer.LastSeen = time.Now()

	fmt.Printf("🏓 Pong from %s - Height: %d, Best Hash: %s\n", 
		peer.Address, pongData.Height, pongData.BestHash[:16])
}

// handleGetBlocks processes block requests
func (mh *MessageHandler) handleGetBlocks(peer *Peer, message NetworkMessage) {
	// For simplicity, send the entire chain
	// In production, this would implement proper block synchronization
	blocksData := BlocksMessage{
		Blocks: mh.node.blockchain.Chain,
	}

	mh.sendMessage(peer, MessageTypeBlocks, blocksData)
}

// handleBlocks processes incoming blocks
func (mh *MessageHandler) handleBlocks(peer *Peer, message NetworkMessage) {
	var blocksData BlocksMessage
	if err := json.Unmarshal(message.Data, &blocksData); err != nil {
		fmt.Printf("❌ Invalid blocks data: %v\n", err)
		return
	}

	fmt.Printf("📦 Received %d blocks from %s\n", len(blocksData.Blocks), peer.Address)

	// Process received blocks
	for _, block := range blocksData.Blocks {
		if mh.node.blockchain.IsValidBlock(block) {
			mh.node.blockchain.AddBlock(block)
			fmt.Printf("✅ Added block %d to chain\n", block.Index)
		}
	}
}

// handleNewBlock processes new block announcements
func (mh *MessageHandler) handleNewBlock(peer *Peer, message NetworkMessage) {
	var newBlockData NewBlockMessage
	if err := json.Unmarshal(message.Data, &newBlockData); err != nil {
		fmt.Printf("❌ Invalid new block data: %v\n", err)
		return
	}

	block := newBlockData.Block
	fmt.Printf("🆕 New block announced from %s: Index=%d, Hash=%s\n", 
		peer.Address, block.Index, block.Hash[:16])

	// Validate and add the block
	if mh.node.blockchain.IsValidBlock(block) {
		mh.node.blockchain.AddBlock(block)
		fmt.Printf("✅ Added new block %d to chain\n", block.Index)
		
		// Broadcast to other peers
		mh.node.BroadcastMessage(rawData)
	} else {
		fmt.Printf("❌ Invalid block received from %s\n", peer.Address)
	}
}

// handleNewTx processes new transaction announcements
func (mh *MessageHandler) handleNewTx(peer *Peer, message NetworkMessage) {
	var newTxData NewTxMessage
	if err := json.Unmarshal(message.Data, &newTxData); err != nil {
		fmt.Printf("❌ Invalid new transaction data: %v\n", err)
		return
	}

	tx := newTxData.Transaction
	fmt.Printf("🆕 New transaction announced from %s: Hash=%s\n", 
		peer.Address, tx.Hash[:16])

	// Validate and add the transaction
	if tx.IsValid() {
		mh.node.blockchain.AddTransaction(tx)
		fmt.Printf("✅ Added transaction to pool: %s\n", tx.Hash[:16])
		
		// Broadcast to other peers
		mh.node.BroadcastMessage(rawData)
	} else {
		fmt.Printf("❌ Invalid transaction received from %s\n", peer.Address)
	}
}

// handleGetPeers processes peer list requests
func (mh *MessageHandler) handleGetPeers(peer *Peer, message NetworkMessage) {
	// Send our peer list
	peers := mh.node.GetPeerList()
	peersData := PeersMessage{
		Peers: peers,
	}

	mh.sendMessage(peer, MessageTypePeers, peersData)
}

// handlePeers processes incoming peer lists
func (mh *MessageHandler) handlePeers(peer *Peer, message NetworkMessage) {
	var peersData PeersMessage
	if err := json.Unmarshal(message.Data, &peersData); err != nil {
		fmt.Printf("❌ Invalid peers data: %v\n", err)
		return
	}

	fmt.Printf("👥 Received %d peers from %s\n", len(peersData.Peers), peer.Address)

	// Connect to new peers
	for _, peerAddr := range peersData.Peers {
		if !mh.node.HasPeer(peerAddr) && peerAddr != mh.node.config.Host {
			go mh.node.connectToNode(peerAddr)
		}
	}
}

// sendMessage sends a message to a peer
func (mh *MessageHandler) sendMessage(peer *Peer, msgType MessageType, data interface{}) {
	message := NetworkMessage{
		Type:      msgType,
		Timestamp: time.Now().Unix(),
		NodeID:    mh.node.config.NodeID,
		Version:   mh.node.config.Version,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("❌ Failed to marshal message data: %v\n", err)
		return
	}
	message.Data = jsonData

	rawMessage, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("❌ Failed to marshal message: %v\n", err)
		return
	}

	if peer.Connected {
		_, err := peer.Conn.Write(rawMessage)
		if err != nil {
			fmt.Printf("❌ Failed to send message to %s: %v\n", peer.Address, err)
		}
	}
}

// GetPeerList returns list of peer addresses
func (n *Node) GetPeerList() []string {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	var peers []string
	for _, peer := range n.peers {
		peers = append(peers, peer.Address)
	}
	return peers
}

// HasPeer checks if we're already connected to a peer
func (n *Node) HasPeer(address string) bool {
	n.peerMutex.RLock()
	defer n.peerMutex.RUnlock()

	for _, peer := range n.peers {
		if peer.Address == address {
			return true
		}
	}
	return false
}