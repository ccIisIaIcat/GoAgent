# MCP集成功能

ConversationManager现已内置Model Context Protocol (MCP) 客户端功能，允许无缝集成外部MCP服务器提供的工具。

## 主要特性

- **组合式设计**：MCP功能直接集成在ConversationManager中
- **自动工具注册**：MCP工具自动注册为ConversationManager函数
- **配置驱动**：支持JSON配置文件管理服务器
- **多传输支持**：支持stdio和TCP传输方式
- **错误隔离**：单个服务器失败不影响其他功能

## 快速开始

### 1. 创建配置文件 (mcp_config.json)
```json
{
  "servers": [
    {
      "name": "file_operations",
      "command": ["python", "servers/file_server.py"],
      "transport": "stdio",
      "env": {
        "PYTHONPATH": "."
      }
    },
    {
      "name": "web_search",
      "command": ["node", "servers/search_server.js"],
      "transport": "stdio"
    }
  ]
}
```

### 2. 初始化ConversationManager并加载MCP配置
```go
// 创建ConversationManager（会自动初始化MCP功能）
cm := ConversationManager.NewConversationManager(manager)

// 加载MCP配置文件
if err := cm.LoadMCPConfig("mcp_config.json"); err != nil {
    log.Printf("加载MCP配置失败: %v", err)
}

// 确保清理资源
defer cm.CloseMCP()
```

### 3. 手动添加服务器（可选）
```go
serverConfig := &ConversationManager.MCPServerConfig{
    Name:      "database",
    Command:   []string{"python", "db_server.py"},
    Transport: "stdio",
}

if err := cm.AddMCPServer(serverConfig); err != nil {
    log.Printf("添加服务器失败: %v", err)
}
```

## API方法

### 配置管理
- `LoadMCPConfig(configPath string) error` - 从文件加载配置
- `LoadMCPConfigFromBytes(data []byte) error` - 从字节数组加载配置
- `CreateMCPConfigTemplate(configPath string) error` - 创建配置模板

### 服务器管理
- `AddMCPServer(config *MCPServerConfig) error` - 添加服务器
- `AddMCPServerFromJSON(jsonStr string) error` - 从JSON添加服务器
- `RemoveMCPServer(serverName string) error` - 移除服务器
- `CloseMCP() error` - 关闭所有连接

### 状态查询
- `GetMCPTools() map[string]*MCPToolInfo` - 获取MCP工具列表
- `GetMCPServerStatus() map[string]interface{}` - 获取服务器状态

## 工具命名规则

MCP工具会自动添加前缀避免命名冲突：
- 原工具名：`read_file`
- 注册后：`mcp_file_operations_read_file`

## 传输方式

### stdio传输
```json
{
  "name": "my_server",
  "command": ["python", "server.py"],
  "transport": "stdio",
  "args": ["--verbose"],
  "env": {
    "DEBUG": "1"
  }
}
```

### TCP传输（计划中）
```json
{
  "name": "tcp_server",
  "address": "localhost:8080",
  "transport": "tcp"
}
```

## 完整使用示例

参见 `example_mcp_usage.go` 文件中的详细示例代码。

## 注意事项

- MCP工具会自动注册到ConversationManager，LLM可直接调用
- 服务器连接失败不会中断其他服务器的连接
- 建议在程序结束时调用 `CloseMCP()` 清理资源