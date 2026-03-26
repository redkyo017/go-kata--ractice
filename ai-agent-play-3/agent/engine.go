package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Agent represents an AI agent that can use tools to complete tasks - ORCHESTRATOR/ENGINE
type Agent struct {
	tools        *ToolRegistry
	llmClient    LLMClient
	maxIteration int
}

// LLMClient represents a language model interface - DECISION MAKER/BRAIN
type LLMClient interface {
	//  GenerateResponse sends a prompt to the LLM and gets back a response
	GenerateResponse(ctx context.Context, prompt string) (string, error)
}

// ToolCall represents a decision by the agent to call a tool
type ToolCall struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
	Thought  string `json:"thought"`
}

// ToolResult holds the result of executing a tool - MEMORY
type ToolResult struct {
	ToolName string
	Output   string
	Error    error
}

func NewAgent(tools *ToolRegistry, llmClient LLMClient) *Agent {
	return &Agent{
		tools:        tools,
		llmClient:    llmClient,
		maxIteration: 10,
	}
}

// Run executes the agent loop
func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	conversationHistory := []string{task}
	iterations := 0

	for iterations < a.maxIteration {
		iterations++
		log.Printf("Iteration %d: Processing task", iterations)

		// Build the prompt with context
		prompt := a.buildPrompt(conversationHistory)

		// Get response from
		response, err := a.llmClient.GenerateResponse(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("LLM error: %w", err)
		}

		log.Printf("LLM response: %v", response)

		// try to parse a tool call from the response
		toolCall, isFinalAnswer := a.parseResponse(response)

		if isFinalAnswer {
			// Agent has provided the final answer
			return response, nil
		}

		if toolCall == nil {
			// Couldn't parse a valid tool call
			conversationHistory = append(conversationHistory, "Invalid response format. Please use JSON format with tool_name, input, and thought fields.")
			continue
		}

		// Execute the tool
		tool := a.tools.Get(toolCall.ToolName)
		if tool == nil {
			conversationHistory = append(conversationHistory, fmt.Sprintf("Tool '%s' not found. Available tools: %v", toolCall.ToolName, a.GetToolNames()))
			continue
		}

		toolResult, err := tool.Execute(ctx, toolCall.Input)
		if err != nil {
			conversationHistory = append(conversationHistory, fmt.Sprintf("Tool execute failed: %v", err))
			continue
		}

		log.Printf("Tool '%s' executed successfully", toolCall.ToolName)

		// Add result to conversation history
		resultMessage := fmt.Sprintf("Tool '%s' returned: %s", toolCall.ToolName, toolResult)
		conversationHistory = append(conversationHistory, resultMessage)
	}

	return "", fmt.Errorf("max iterations reached without completing the task")
}

func (a *Agent) buildPrompt(history []string) string {
	toolDescriptions := a.getToolDescriptions()

	prompt := `You are an AI agent that can use tools to complete tasks.
	Avalable tools:
	` + toolDescriptions + `
	Instructions:
	1. If you need information, use the appropriate tool
	2. After using a tool, wait for the result and use it to continue
	3. When you have enough information to answer the original question, provide the final answer
	4. Always respond in JSON format with this structure:
	{
		"tool_name": "name_of_tool",
		"input": "input_for_tool",
		"thought": "Your reasoning"
	}
	5. When providing the final answer, use this format:
	{
		"final_answer": "Your complete answer here"
	}

	Conversation history:
	`
	for i, entry := range history {
		prompt += fmt.Sprintf("%d. %s\n", i+1, entry)
	}

	prompt += "\nWhat should be your next actions?"

	return prompt
}

func (a *Agent) getToolDescriptions() string {
	var descriptions string
	for _, tool := range a.tools.List() {
		descriptions += fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description())
	}
	return descriptions
}

func (a *Agent) GetToolNames() []string {
	var names []string
	for _, tool := range a.tools.List() {
		names = append(names, tool.Name())
	}
	return names
}

func (a *Agent) parseResponse(response string) (*ToolCall, bool) {
	// Try to find JSON in the response
	var jsonStr string

	// Look for JSON object in the response
	startIdx := -1
	for i := 0; i < len(response); i++ {
		if response[i] == '{' {
			startIdx = i
			break
		}
	}
	if startIdx == -1 {
		return nil, false
	}

	// Extract JSON
	braceCount := 0
	endIdx := -1

	for i := startIdx; i < len(response); i++ {
		if response[i] == '{' {
			braceCount++
		} else if response[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}
	if endIdx == -1 {
		return nil, false
	}
	jsonStr = response[startIdx:endIdx]

	// Check if this is a final answer
	var finalAnswerCheck map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &finalAnswerCheck); err == nil {
		if _, hasFinalAnswer := finalAnswerCheck["final_answer"]; hasFinalAnswer {
			return nil, true
		}
	}

	// Try to parse as tool call
	var toolCall ToolCall
	if err := json.Unmarshal([]byte(jsonStr), &toolCall); err != nil {
		return nil, false
	}

	if toolCall.ToolName == "" || toolCall.Input == "" {
		return nil, false
	}

	return &toolCall, false
}
