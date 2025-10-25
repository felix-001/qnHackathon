package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/felix-001/qnHackathon/internal/model"
)

type BinaryService struct {
	binaries map[string]*model.BinaryInfo
	nodes    map[string]*model.NodeInfo
	mu       sync.RWMutex
	binDir   string
}

func NewBinaryService(binDir string) *BinaryService {
	if binDir == "" {
		binDir = "./bins"
	}

	os.MkdirAll(binDir, 0755)

	return &BinaryService{
		binaries: make(map[string]*model.BinaryInfo),
		nodes:    make(map[string]*model.NodeInfo),
		binDir:   binDir,
	}
}

func (s *BinaryService) GetBinaryHash(name string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if bin, exists := s.binaries[name]; exists {
		return bin.Hash, nil
	}

	binPath := filepath.Join(s.binDir, name)
	hash, err := s.calculateFileHash(binPath)
	if err != nil {
		return "", err
	}

	s.mu.RUnlock()
	s.mu.Lock()
	s.binaries[name] = &model.BinaryInfo{
		Name:      name,
		Version:   "latest",
		Hash:      hash,
		UpdatedAt: time.Now(),
	}
	s.mu.Unlock()
	s.mu.RLock()

	return hash, nil
}

func (s *BinaryService) UpdateBinaryHash(name, hash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if bin, exists := s.binaries[name]; exists {
		bin.Hash = hash
		bin.UpdatedAt = time.Now()
	} else {
		s.binaries[name] = &model.BinaryInfo{
			Name:      name,
			Version:   "latest",
			Hash:      hash,
			UpdatedAt: time.Now(),
		}
	}

	return nil
}

func (s *BinaryService) GetNodeInfo(nodeName string) (*model.NodeInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[nodeName]
	return node, exists
}

func (s *BinaryService) UpdateNodeInfo(node *model.NodeInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node.LastSeen = time.Now()
	s.nodes[node.Name] = node

	return nil
}

func (s *BinaryService) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s *BinaryService) GetBinaryPath(name string) string {
	return filepath.Join(s.binDir, name)
}
