package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigService struct {
	configDir string
}

func NewConfigService() *ConfigService {
	configDir := "./configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Failed to create config directory: %v\n", err)
	}
	return &ConfigService{
		configDir: configDir,
	}
}

func (s *ConfigService) GetConfig(projectName string) (map[string]interface{}, error) {
	filePath := filepath.Join(s.configDir, projectName+".json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

func (s *ConfigService) SaveConfig(projectName string, content string) error {
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	filePath := filepath.Join(s.configDir, projectName+".json")
	
	formattedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %v", err)
	}

	if err := os.WriteFile(filePath, formattedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
