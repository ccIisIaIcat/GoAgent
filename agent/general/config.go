package general

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// APIConfig 单个API配置
type APIConfig struct {
	BaseUrl string `yaml:"BaseUrl"`
	APIKey  string `yaml:"APIKey"`
	Model   string `yaml:"Model,omitempty"` // 可选的模型名称
}

// LLMConfig 完整的LLM配置
type LLMConfig struct {
	AgentAPIKey struct {
		OpenAI    APIConfig `yaml:"OpenAI"`
		Anthropic APIConfig `yaml:"Anthropic"`
		DeepSeek  APIConfig `yaml:"DeepSeek"`
		GoogleKey APIConfig `yaml:"GoogleKey"`
		Qwen      APIConfig `yaml:"Qwen"`
	} `yaml:"AgentAPIKey"`
}

// LoadConfig 从YAML文件加载配置
func LoadConfig(filename string) (*LLMConfig, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s does not exist", filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML
	var config LLMConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// getDefaultModel 获取默认模型名称
func getDefaultModel(provider Provider) string {
	switch provider {
	case ProviderOpenAI:
		return "gpt-4o"
	case ProviderAnthropic:
		return "claude-3-5-sonnet-20241022"
	case ProviderDeepSeek:
		return "deepseek-chat"
	case ProviderGoogle:
		return "gemini-pro"
	case ProviderQwen:
		return "qwen-plus"
	default:
		return ""
	}
}

// ToProviderConfig 将配置转换为ProviderConfig切片
func (c *LLMConfig) ToProviderConfigs() []*ProviderConfig {
	var configs []*ProviderConfig

	// OpenAI配置
	if c.AgentAPIKey.OpenAI.APIKey != "" {
		model := c.AgentAPIKey.OpenAI.Model
		if model == "" {
			model = getDefaultModel(ProviderOpenAI)
		}
		configs = append(configs, &ProviderConfig{
			Provider: ProviderOpenAI,
			APIKey:   c.AgentAPIKey.OpenAI.APIKey,
			BaseURL:  c.AgentAPIKey.OpenAI.BaseUrl,
			Model:    model,
		})
	}

	// Anthropic配置
	if c.AgentAPIKey.Anthropic.APIKey != "" {
		model := c.AgentAPIKey.Anthropic.Model
		if model == "" {
			model = getDefaultModel(ProviderAnthropic)
		}
		configs = append(configs, &ProviderConfig{
			Provider: ProviderAnthropic,
			APIKey:   c.AgentAPIKey.Anthropic.APIKey,
			BaseURL:  c.AgentAPIKey.Anthropic.BaseUrl,
			Model:    model,
		})
	}

	// DeepSeek配置
	if c.AgentAPIKey.DeepSeek.APIKey != "" {
		model := c.AgentAPIKey.DeepSeek.Model
		if model == "" {
			model = getDefaultModel(ProviderDeepSeek)
		}
		configs = append(configs, &ProviderConfig{
			Provider: ProviderDeepSeek,
			APIKey:   c.AgentAPIKey.DeepSeek.APIKey,
			BaseURL:  c.AgentAPIKey.DeepSeek.BaseUrl,
			Model:    model,
		})
	}

	// Google配置
	if c.AgentAPIKey.GoogleKey.APIKey != "" {
		model := c.AgentAPIKey.GoogleKey.Model
		if model == "" {
			model = getDefaultModel(ProviderGoogle)
		}
		configs = append(configs, &ProviderConfig{
			Provider: ProviderGoogle,
			APIKey:   c.AgentAPIKey.GoogleKey.APIKey,
			BaseURL:  c.AgentAPIKey.GoogleKey.BaseUrl,
			Model:    model,
		})
	}

	// Qwen(阿里云)配置
	if c.AgentAPIKey.Qwen.APIKey != "" {
		model := c.AgentAPIKey.Qwen.Model
		if model == "" {
			model = getDefaultModel(ProviderQwen)
		}
		configs = append(configs, &ProviderConfig{
			Provider: ProviderQwen,
			APIKey:   c.AgentAPIKey.Qwen.APIKey,
			BaseURL:  c.AgentAPIKey.Qwen.BaseUrl,
			Model:    model,
		})
	}

	return configs
}

// PrintConfig 打印配置信息（隐藏API密钥）
func (c *LLMConfig) PrintConfig() {
	fmt.Println("=== LLM Configuration ===")

	if c.AgentAPIKey.OpenAI.APIKey != "" {
		model := c.AgentAPIKey.OpenAI.Model
		if model == "" {
			model = getDefaultModel(ProviderOpenAI)
		}
		fmt.Printf("OpenAI: %s | Model: %s | Key: %s...\n",
			c.AgentAPIKey.OpenAI.BaseUrl,
			model,
			maskAPIKey(c.AgentAPIKey.OpenAI.APIKey))
	}

	if c.AgentAPIKey.Anthropic.APIKey != "" {
		model := c.AgentAPIKey.Anthropic.Model
		if model == "" {
			model = getDefaultModel(ProviderAnthropic)
		}
		fmt.Printf("Anthropic: %s | Model: %s | Key: %s...\n",
			c.AgentAPIKey.Anthropic.BaseUrl,
			model,
			maskAPIKey(c.AgentAPIKey.Anthropic.APIKey))
	}

	if c.AgentAPIKey.DeepSeek.APIKey != "" {
		model := c.AgentAPIKey.DeepSeek.Model
		if model == "" {
			model = getDefaultModel(ProviderDeepSeek)
		}
		fmt.Printf("DeepSeek: %s | Model: %s | Key: %s...\n",
			c.AgentAPIKey.DeepSeek.BaseUrl,
			model,
			maskAPIKey(c.AgentAPIKey.DeepSeek.APIKey))
	}

	if c.AgentAPIKey.GoogleKey.APIKey != "" {
		model := c.AgentAPIKey.GoogleKey.Model
		if model == "" {
			model = getDefaultModel(ProviderGoogle)
		}
		fmt.Printf("Google: %s | Model: %s | Key: %s...\n",
			c.AgentAPIKey.GoogleKey.BaseUrl,
			model,
			maskAPIKey(c.AgentAPIKey.GoogleKey.APIKey))
	}

	if c.AgentAPIKey.Qwen.APIKey != "" {
		model := c.AgentAPIKey.Qwen.Model
		if model == "" {
			model = getDefaultModel(ProviderQwen)
		}
		fmt.Printf("Qwen(Ali): %s | Model: %s | Key: %s...\n",
			c.AgentAPIKey.Qwen.BaseUrl,
			model,
			maskAPIKey(c.AgentAPIKey.Qwen.APIKey))
	}

	fmt.Println("========================")
}

// maskAPIKey 遮盖API密钥显示
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return apiKey
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}
