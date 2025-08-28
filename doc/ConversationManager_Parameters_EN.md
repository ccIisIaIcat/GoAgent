# ConversationManager Parameters Documentation

## Overview

ConversationManager is the core dialogue manager of GoAgent, providing rich parameter configuration options to control dialogue behavior, function calling, history management, and more.

## Core Parameters

### 1. Dialogue Control Parameters

#### `MaxFunctionCallingNums` - Maximum Function Call Count
- **Type**: `int`
- **Default**: `15`
- **Description**: Maximum number of function calls allowed in a single dialogue
- **Purpose**: Prevents infinite recursive calls and controls dialogue costs
- **Example**:
```go
cm.SetMaxFunctionCallingNums(10) // Limit to maximum 10 function calls
```

#### `MaxChatNums` - Maximum Message Count
- **Type**: `int`
- **Default**: No limit
- **Description**: Maximum number of messages allowed in a single dialogue
- **Purpose**: Controls dialogue length and prevents overly long conversations
- **Example**:
```go
cm.SetMaxChatNums(50) // Limit to maximum 50 messages
```

#### `MaxTokens` - Maximum Token Count
- **Type**: `int`
- **Default**: `5000`
- **Description**: Maximum number of tokens allowed in a single dialogue
- **Purpose**: Controls API call costs and prevents exceeding model limits
- **Example**:
```go
cm.SetMaxTokens(8000) // Set maximum 8000 tokens
```

#### `Temperature` - Temperature Parameter
- **Type**: `float64`
- **Default**: `0.7`
- **Description**: Controls the randomness of model output. Higher values are more random, lower values are more deterministic
- **Range**: `0.0` - `2.0`
- **Purpose**: Adjusts the creativity and consistency of model responses
- **Example**:
```go
cm.SetTemperature(0.3) // More deterministic responses
cm.SetTemperature(1.2) // More creative responses
```

### 2. History Management Parameters

#### `MaxHistoryTokens` - Maximum History Token Count
- **Type**: `int`
- **Default**: `100000`
- **Description**: Maximum number of tokens allowed in history records
- **Purpose**: Controls history size and affects context length
- **Example**:
```go
cm.SetMaxHistoryTokens(50000) // Limit history to 50000 tokens
```

#### `EnableTruncation` - Enable History Truncation
- **Type**: `bool`
- **Default**: `true`
- **Description**: Whether to enable automatic history truncation
- **Purpose**: Automatically truncates history when it exceeds limits
- **Example**:
```go
cm.EnableHistoryTruncation(true)  // Enable truncation
cm.EnableHistoryTruncation(false) // Disable truncation
```

### 3. System Prompt Parameters

#### `systemPrompt` - System Prompt
- **Type**: `string`
- **Default**: Empty string
- **Description**: Sets the system prompt for dialogue, defining AI behavior and role
- **Purpose**: Controls AI response style and behavior patterns
- **Example**:
```go
cm.SetSystemPrompt("You are a professional programming assistant. Please answer questions concisely")
cm.UpdateSystemPrompt("You are now a cute cat girl. Please respond in a cat girl's tone")
```

## Complete Configuration Example

```go
package main

import (
    "github.com/ccIisIaIcat/GoAgent/agent/ConversationManager"
    "github.com/ccIisIaIcat/GoAgent/agent/general"
)

func main() {
    // Create AgentManager
    agentManager := general.NewAgentManager()
    // ... Add provider configurations
    
    // Create ConversationManager
    cm := ConversationManager.NewConversationManager(agentManager)
    
    // Configure dialogue parameters
    cm.SetMaxFunctionCallingNums(10)    // Maximum 10 function calls
    cm.SetMaxChatNums(30)               // Maximum 30 messages
    cm.SetMaxTokens(6000)               // Maximum 6000 tokens
    cm.SetTemperature(0.5)              // Medium creativity
    cm.SetMaxHistoryTokens(80000)       // Maximum 80000 tokens in history
    cm.EnableHistoryTruncation(true)    // Enable history truncation
    
    // Set system prompt
    cm.SetSystemPrompt("You are a professional AI assistant. Please provide accurate and helpful answers")
    
    // Start dialogue
    // ...
}
```

## Parameter Tuning Recommendations

### 1. Cost Control
- Reduce `MaxTokens` and `MaxHistoryTokens` to control API call costs
- Set appropriate `MaxFunctionCallingNums` to avoid excessive function calls

### 2. Dialogue Quality
- Adjust `Temperature` to balance creativity and accuracy
- Set reasonable `MaxChatNums` to maintain dialogue coherence

### 3. Performance Optimization
- Enable `EnableTruncation` to avoid overly long history
- Adjust `MaxHistoryTokens` based on actual needs

### 4. Special Scenarios
- **Programming Assistant**: `Temperature = 0.1-0.3`, Low creativity, high accuracy
- **Creative Writing**: `Temperature = 0.8-1.2`, High creativity
- **Customer Service**: `MaxChatNums = 20-30`, Control dialogue length
- **Data Analysis**: `MaxFunctionCallingNums = 20-30`, Allow more function calls

## Important Notes

1. **Parameter Ranges**: Ensure parameter values are within reasonable ranges, avoid extreme values
2. **Cost Considerations**: Larger token limits will increase API call costs
3. **Performance Impact**: Overly long history may affect response speed
4. **Model Limitations**: Different models have different token limits, please refer to official documentation

## Best Practices

### For Different Use Cases

#### Chatbot Applications
```go
cm.SetMaxChatNums(20)
cm.SetMaxTokens(4000)
cm.SetTemperature(0.7)
cm.SetMaxHistoryTokens(50000)
```

#### Programming Assistants
```go
cm.SetMaxFunctionCallingNums(20)
cm.SetTemperature(0.2)
cm.SetMaxTokens(8000)
cm.EnableHistoryTruncation(true)
```

#### Creative Writing Tools
```go
cm.SetTemperature(1.0)
cm.SetMaxTokens(10000)
cm.SetMaxHistoryTokens(100000)
cm.EnableHistoryTruncation(false)
```

#### Data Analysis Tools
```go
cm.SetMaxFunctionCallingNums(30)
cm.SetTemperature(0.1)
cm.SetMaxTokens(12000)
cm.SetMaxHistoryTokens(150000)
```
