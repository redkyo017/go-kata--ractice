package main

import (
	"ai-agent-play-3/agent"
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()

	// set up tools
	toolRegistry := agent.NewToolRegistry()
	toolRegistry.Register(&agent.CalculatorTool{})
	toolRegistry.Register(&agent.WebSearchTool{})
	toolRegistry.Register(&agent.FileReaderTool{})

	// initialize LLM client - using Ollama locally
	llmClient := agent.NewOllamaClient("http://localhost:11434", "mistral")

	// Creat agent
	agentInstance := agent.NewAgent(toolRegistry, llmClient)

	// Run the agent on a task
	task := "Find out what 45 multiplied by 12 is, and then tell me if it's greater than 500"

	log.Println("Start agent wit task:", task)

	result, err := agentInstance.Run(ctx, task)
	if err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("Final Result:")
	fmt.Println(result)
}
