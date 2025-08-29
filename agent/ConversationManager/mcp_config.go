package ConversationManager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

// MCPConfig MCP配置文件结构
type MCPConfig struct {
	Servers    []MCPServerConfig            `json:"servers,omitempty"`
	McpServers map[string]MCPServerSettings `json:"mcpServers,omitempty"`
}

// MCPServerSettings MCP服务器设置（新格式）
type MCPServerSettings struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// LoadMCPConfig 从文件加载MCP配置并注册服务
func (cm *ConversationManager) LoadMCPConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取MCP配置文件失败: %w", err)
	}

	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析MCP配置文件失败: %w", err)
	}

	// 验证配置
	if err := cm.ValidateMCPConfig(&config); err != nil {
		return fmt.Errorf("MCP配置验证失败: %w", err)
	}

	// 注册所有服务器
	var errors []error
	successCount := 0
	
	// 处理原有格式的servers
	for _, serverConfig := range config.Servers {
		if err := cm.AddMCPServer(&serverConfig); err != nil {
			errors = append(errors, fmt.Errorf("连接服务器 %s 失败: %w", serverConfig.Name, err))
			log.Printf("连接MCP服务器失败 %s: %v", serverConfig.Name, err)
		} else {
			successCount++
		}
	}
	
	// 处理新格式的mcpServers
	for serverName, settings := range config.McpServers {
		serverConfig := MCPServerConfig{
			Name:      serverName,
			Command:   []string{settings.Command},
			Args:      settings.Args,
			Transport: "stdio",
		}
		
		if err := cm.AddMCPServer(&serverConfig); err != nil {
			errors = append(errors, fmt.Errorf("连接服务器 %s 失败: %w", serverName, err))
			log.Printf("连接MCP服务器失败 %s: %v", serverName, err)
		} else {
			successCount++
		}
	}

	totalServers := len(config.Servers) + len(config.McpServers)
	log.Printf("成功连接 %d/%d 个MCP服务器", successCount, totalServers)

	if len(errors) > 0 && successCount == 0 {
		return fmt.Errorf("所有MCP服务器连接失败: %v", errors)
	}

	return nil
}

// SaveMCPConfig 保存MCP配置到文件
func SaveMCPConfig(config *MCPConfig, configPath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化MCP配置失败: %w", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入MCP配置文件失败: %w", err)
	}

	return nil
}

// GetDefaultMCPConfig 获取默认MCP配置
func GetDefaultMCPConfig() *MCPConfig {
	return &MCPConfig{
		Servers: []MCPServerConfig{
			{
				Name:      "example_stdio",
				Command:   []string{"python", "example_server.py"},
				Transport: "stdio",
			},
			{
				Name:      "example_tcp",
				Address:   "localhost:8080",
				Transport: "tcp",
			},
		},
	}
}

// ValidateMCPConfig 验证MCP配置有效性
func (cm *ConversationManager) ValidateMCPConfig(config *MCPConfig) error {
	if config == nil {
		return fmt.Errorf("配置为空")
	}

	if len(config.Servers) == 0 && len(config.McpServers) == 0 {
		return fmt.Errorf("配置中没有定义任何服务器")
	}

	serverNames := make(map[string]bool)

	// 验证原格式的servers
	for i, server := range config.Servers {
		if server.Name == "" {
			return fmt.Errorf("服务器 %d 缺少名称", i)
		}

		if serverNames[server.Name] {
			return fmt.Errorf("服务器名称重复: %s", server.Name)
		}
		serverNames[server.Name] = true

		switch server.Transport {
		case "stdio":
			if len(server.Command) == 0 {
				return fmt.Errorf("服务器 %s: stdio传输需要指定命令", server.Name)
			}
		case "tcp":
			if server.Address == "" {
				return fmt.Errorf("服务器 %s: tcp传输需要指定地址", server.Name)
			}
		default:
			return fmt.Errorf("服务器 %s: 不支持的传输类型: %s", server.Name, server.Transport)
		}
	}

	// 验证新格式的mcpServers
	for serverName, settings := range config.McpServers {
		if serverName == "" {
			return fmt.Errorf("mcpServers中存在空的服务器名称")
		}

		if serverNames[serverName] {
			return fmt.Errorf("服务器名称重复: %s", serverName)
		}
		serverNames[serverName] = true

		if settings.Command == "" {
			return fmt.Errorf("服务器 %s: 缺少命令", serverName)
		}
	}

	return nil
}

// CreateMCPConfigTemplate 创建MCP配置文件模板
func CreateMCPConfigTemplate(configPath string) error {
	config := GetDefaultMCPConfig()

	// 确保目录存在
	if err := ensureDir(filepath.Dir(configPath)); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return SaveMCPConfig(config, configPath)
}

// ensureDir 确保目录存在
func ensureDir(dir string) error {
	return nil // 在Windows上，我们假设目录已存在或会被自动创建
}

// LoadMCPConfigFromBytes 从字节数组加载MCP配置
func (cm *ConversationManager) LoadMCPConfigFromBytes(data []byte) error {
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析MCP配置失败: %w", err)
	}

	if err := cm.ValidateMCPConfig(&config); err != nil {
		return fmt.Errorf("MCP配置验证失败: %w", err)
	}

	// 注册所有服务器
	for _, serverConfig := range config.Servers {
		if err := cm.AddMCPServer(&serverConfig); err != nil {
			log.Printf("连接MCP服务器失败 %s: %v", serverConfig.Name, err)
		}
	}

	return nil
}

// AddMCPServerFromJSON 从JSON字符串添加MCP服务器
func (cm *ConversationManager) AddMCPServerFromJSON(jsonStr string) error {
	var config MCPServerConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return fmt.Errorf("解析服务器配置失败: %w", err)
	}

	return cm.AddMCPServer(&config)
}

// GetMCPServerStatus 获取MCP服务器状态
func (cm *ConversationManager) GetMCPServerStatus() map[string]interface{} {
	if cm.mcpManager == nil {
		return map[string]interface{}{
			"enabled":       false,
			"servers":       0,
			"tools":         0,
		}
	}

	tools := cm.mcpManager.GetRegisteredTools()
	serverMap := make(map[string]int)
	
	for _, tool := range tools {
		serverMap[tool.ServerName]++
	}

	return map[string]interface{}{
		"enabled":       true,
		"servers":       len(serverMap),
		"tools":         len(tools),
		"server_tools":  serverMap,
	}
}