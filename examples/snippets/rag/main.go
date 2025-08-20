package main

import (
	"log"

	"github.com/oopslink/agent-go/pkg/core/agent/behavior_patterns"
	"github.com/oopslink/agent-go/pkg/core/knowledge"
	_ "github.com/oopslink/agent-go/pkg/support/llms/anthropic"
	_ "github.com/oopslink/agent-go/pkg/support/llms/gemini"
	_ "github.com/oopslink/agent-go/pkg/support/llms/openai"

	u "github.com/oopslink/agent-go-snippets/utils"
)

func main() {
	// parse arguments
	question, apiKey, provider, modelName := u.ParseArgs("What is machine learning?")

	// create RAG agent
	// 1. create knowledge base with sample data
	knowledgeBase, err := createSampleKnowledgeBase()
	if err != nil {
		log.Fatalf("failed to create knowledge base: %v", err)
	}
	// 2. create RAG behavior pattern with knowledge base
	behavior, err := behavior_patterns.NewRAGPatternWithKnowledgeBases([]knowledge.KnowledgeBase{knowledgeBase})
	if err != nil {
		log.Fatalf("failed to create RAG behavior pattern: %v", err)
	}
	// 3. create RAG Agent instance
	ragAgent, err := u.CreateAgent(
		"rag-demo",
		"You are a helpful AI assistant",
		apiKey, provider, modelName,
		behavior, nil, nil, false)
	if err != nil {
		log.Fatalf("Failed to create RAG agent: %v", err)
	}

	// run single conversation
	if err = u.RunSingleConversation(ragAgent,
		"🤖 Agent searching knowledge base and generating response...", question); err != nil {
		log.Fatalf("Failed to run conversation: %v", err)
	}
}
