package service

import (
	"sync"
	"time"
)

type Node struct {
	NodeID          string    `json:"node_id"`
	CPUArch         string    `json:"cpu_arch"`
	OSRelease       string    `json:"os_release"`
	NodeName        string    `json:"node_name"`
	BinProxyVersion string    `json:"bin_proxy_version"`
	LastSeen        time.Time `json:"last_seen"`
}

type Bin struct {
	BinName   string `json:"bin_name"`
	SHA256Sum string `json:"sha256sum"`
	Version   string `json:"version"`
}

type NodeBin struct {
	SHA256Sum string    `json:"sha256sum"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Progress struct {
	NodeName       string    `json:"nodeName"`
	BinName        string    `json:"binName"`
	TargetHash     string    `json:"targetHash"`
	ProcessingTime *int      `json:"processingTime,omitempty"`
	Status         string    `json:"status"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BinService struct {
	nodes        map[string]*Node
	bins         map[string]*Bin
	nodeBins     map[string]map[string]*NodeBin
	progressData map[string]*Progress
	mu           sync.RWMutex
}

func NewBinService() *BinService {
	return &BinService{
		nodes:        make(map[string]*Node),
		bins:         initBins(),
		nodeBins:     make(map[string]map[string]*NodeBin),
		progressData: make(map[string]*Progress),
	}
}

func initBins() map[string]*Bin {
	return map[string]*Bin{
		"bin1": {
			BinName:   "bin1",
			SHA256Sum: "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
			Version:   "latest",
		},
		"bin2": {
			BinName:   "bin2",
			SHA256Sum: "1234abcd5678901234abcd5678901234abcd5678901234abcd5678901234abcd",
			Version:   "1.0.0",
		},
		"bin3": {
			BinName:   "bin3",
			SHA256Sum: "1234abcd5678901234abcd5678901234abcd5678901234abcd5678901234abc5",
			Version:   "latest",
		},
	}
}

func (s *BinService) GetNode(nodeID string) (*Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	node, ok := s.nodes[nodeID]
	return node, ok
}

func (s *BinService) RegisterNode(nodeID, cpuArch, osRelease, nodeName, binProxyVersion string) *Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	node := &Node{
		NodeID:          nodeID,
		CPUArch:         cpuArch,
		OSRelease:       osRelease,
		NodeName:        nodeName,
		BinProxyVersion: binProxyVersion,
		LastSeen:        time.Now().UTC(),
	}
	s.nodes[nodeID] = node
	return node
}

func (s *BinService) GetBin(binName string) (*Bin, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bin, ok := s.bins[binName]
	return bin, ok
}

func (s *BinService) UpdateNodeBin(nodeID, binName, sha256sum string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nodeBins[nodeID] == nil {
		s.nodeBins[nodeID] = make(map[string]*NodeBin)
	}

	s.nodeBins[nodeID][binName] = &NodeBin{
		SHA256Sum: sha256sum,
		UpdatedAt: time.Now().UTC(),
	}
}

func (s *BinService) RecordProgress(nodeName, binName, targetHash, status string, processingTime *int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	progressKey := nodeName + ":" + binName + ":" + targetHash
	s.progressData[progressKey] = &Progress{
		NodeName:       nodeName,
		BinName:        binName,
		TargetHash:     targetHash,
		ProcessingTime: processingTime,
		Status:         status,
		UpdatedAt:      time.Now().UTC(),
	}
}

func (s *BinService) GetNodesCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.nodes)
}

func (s *BinService) GetBinsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.bins)
}
