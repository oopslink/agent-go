# RAG (Retrieval-Augmented Generation) Example

This example demonstrates how to use the RAG behavior pattern in agent-go to create an AI agent that can retrieve knowledge from a knowledge base and use it to answer questions.

## Features

- **Knowledge Retrieval**: The agent searches through a knowledge base to find relevant information
- **Context-Aware Responses**: Uses retrieved knowledge to provide accurate and comprehensive answers
- **Streaming Responses**: Real-time response generation with streaming output
- **Multiple LLM Providers**: Support for OpenAI, Anthropic, and Gemini models

## Usage

1. Set your API key:
   ```bash
   export API_KEY="your-api-key-here"
   ```

2. Optionally set the provider and model:
   ```bash
   export PROVIDER="openai"  # or "anthropic", "gemini"
   export MODEL="gpt-4o-mini"  # or any other model
   ```

3. Run the example:
   ```bash
   go run *.go "What is machine learning?"
   ```

## Example Questions

Try these example questions to see the RAG pattern in action:

- "What is machine learning?"
- "Explain deep learning"
- "What is NLP?"
- "How does artificial intelligence work?"

## How It Works

1. **Knowledge Base Setup**: The example creates a simple in-memory knowledge base with educational content about AI topics
2. **Query Processing**: When a question is asked, the RAG pattern searches the knowledge base for relevant information
3. **Context Generation**: Retrieved knowledge is formatted and added to the conversation context
4. **Response Generation**: The LLM generates a response based on the retrieved knowledge and the user's question

## Architecture

The RAG pattern follows this workflow:

1. **Knowledge Retrieval**: Searches knowledge bases for relevant content
2. **Prompt Engineering**: Formats retrieved knowledge with the user's question
3. **LLM Generation**: Uses the LLM to generate a response based on the retrieved knowledge
4. **Response Streaming**: Streams the response back to the user

## Customization

To use your own knowledge base:

1. Implement the `knowledge.KnowledgeBase` interface
2. Pass your knowledge base to `NewRAGPatternWithKnowledgeBases()`
3. Update the agent context with your knowledge base

## Dependencies

- agent-go core packages
- LLM provider SDKs (OpenAI, Anthropic, Gemini)
- Event bus for communication
- Memory system for conversation history 