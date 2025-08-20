# Chat Demo with Knowledge Tool

This demo showcases the Agent-Go framework with an integrated knowledge retrieval tool that loads knowledge bases from JSON data and uses real vector database and embedder implementations.

## Features

- **Knowledge Tool Integration**: Automatically loads knowledge bases from `knowledges/data.json`
- **Real Vector Database**: Supports Milvus vector database for efficient similarity search
- **Real Embedder**: Uses LLM-based embedding providers (OpenAI, Anthropic, etc.)
- **Configuration-Based**: All settings configured via JSON configuration file
- **Automatic Tool Calling**: Tools are called automatically without manual confirmation
- **Multiple Knowledge Domains**: Supports Astronomy, Biology, Physics, and History domains
- **Smart Search**: Can search across all domains or specify particular domains

## Knowledge Bases

The demo includes knowledge bases for:

- **Astronomy**: Facts about celestial bodies, space phenomena, and the universe
- **Biology**: Facts about living organisms, evolution, and life processes
- **Physics**: Fundamental laws and phenomena governing matter, energy, and the universe
- **History**: Significant events, civilizations, and developments throughout human history

## Configuration

The application uses a JSON configuration file to set up all components:

### Configuration File Structure

```json
{
  "provider": "openai",
  "api_key": "your-openai-api-key-here",
  "base_url": "https://api.openai.com/v1",
  "model_name": "gpt-4o-mini",
  "openai_compatibility": false,
  "system_prompt": "You are a helpful AI assistant with access to a knowledge base...",
  
  "vector_db": {
    "type": "milvus",
    "endpoint": "localhost:19530",
    "username": "",
    "password": "",
    "timeout": 15
  },
  
  "embedder": {
    "type": "llm",
    "provider": "openai",
    "api_key": "your-openai-api-key-here",
    "base_url": "https://api.openai.com/v1",
    "model": "text-embedding-3-small"
  }
}
```

### Configuration Options

#### Main Configuration
- `provider`: AI model provider (openai, anthropic, gemini)
- `api_key`: API key for the chat provider
- `base_url`: Base URL for the API
- `model_name`: Model name for chat
- `openai_compatibility`: Enable OpenAI compatibility mode
- `system_prompt`: Custom system prompt

#### Vector Database Configuration
- `type`: Vector database type (milvus)
- `endpoint`: Vector database endpoint (e.g., localhost:19530)
- `username`: Vector database username (optional)
- `password`: Vector database password (optional)
- `timeout`: Connection timeout in seconds

#### Embedder Configuration
- `type`: Embedder type (llm)
- `provider`: Embedding provider (openai, anthropic, etc.)
- `api_key`: API key for the embedder provider
- `base_url`: Base URL for the embedder API
- `model`: Embedding model name (e.g., text-embedding-3-small)

## Usage

### Running the Demo

```bash
# Using default config file (config.json)
./demo chat

# Using custom config file
./demo chat --config my-config.json

# Using relative path
./demo chat --config ./config/config.json
```

### Example Configuration Files

#### OpenAI with Milvus
```json
{
  "provider": "openai",
  "api_key": "sk-your-openai-key",
  "base_url": "https://api.openai.com/v1",
  "model_name": "gpt-4o-mini",
  "openai_compatibility": false,
  "system_prompt": "You are a helpful AI assistant with access to a knowledge base...",
  
  "vector_db": {
    "type": "milvus",
    "endpoint": "localhost:19530",
    "username": "",
    "password": "",
    "timeout": 15
  },
  
  "embedder": {
    "type": "llm",
    "provider": "openai",
    "api_key": "sk-your-openai-key",
    "base_url": "https://api.openai.com/v1",
    "model": "text-embedding-3-small"
  }
}
```

#### Anthropic with Milvus
```json
{
  "provider": "anthropic",
  "api_key": "sk-ant-your-anthropic-key",
  "base_url": "https://api.anthropic.com",
  "model_name": "claude-3.5-sonnet",
  "openai_compatibility": false,
  
  "vector_db": {
    "type": "milvus",
    "endpoint": "localhost:19530"
  },
  
  "embedder": {
    "type": "llm",
    "provider": "openai",
    "api_key": "sk-your-openai-key",
    "model": "text-embedding-3-small"
  }
}
```

### Example Interactions

The AI assistant can now answer questions using the knowledge base with real vector search:

```
User: Tell me about black holes
Assistant: Let me search the knowledge base for information about black holes...

[Tool Call: knowledge_retriever]
- Query: black holes
- Domains: [Astronomy]

Result: Black holes don't actually suck things in - objects must cross the event horizon to be trapped by the immense gravitational field

Based on the knowledge base, black holes don't actually "suck" things in as commonly portrayed. Instead, objects must cross the event horizon to be trapped by the immense gravitational field...
```

## Knowledge Tool Features

### Real Vector Search
The knowledge tool now uses real vector database (Milvus) for efficient similarity search:
- **Semantic Search**: Finds relevant documents based on meaning, not just keywords
- **Scalable**: Can handle large knowledge bases efficiently
- **Configurable**: Supports different embedding models and vector databases

### Automatic Domain Selection
The AI can automatically select relevant domains based on the question:
- Astronomy questions → Astronomy domain
- Biology questions → Biology domain
- Physics questions → Physics domain
- History questions → History domain

### Search Parameters
The knowledge tool supports various search parameters:
- `query`: The search query
- `max_results`: Maximum number of results (default: 10)
- `score_threshold`: Minimum similarity score (default: 0.0)
- `domains`: Filter by specific domains
- `max_bases`: Maximum number of knowledge bases to search (-1 for all)

### Tool Descriptor
The tool automatically generates descriptions based on available knowledge bases:
```
"Filter by knowledge base domains. Available domains: [Astronomy Biology Physics History]"
```

## File Structure

```
demo/apps/chat/
├── chat.go                 # Main chat application
├── config/
│   ├── config.go          # Configuration structures
│   └── config.json        # Sample configuration
├── knowledges/
│   ├── data.json          # Knowledge base data
│   ├── loader.go          # Knowledge base loader
│   └── loader_test.go     # Tests for loader
└── ui/
    └── ui.go              # UI components
```

## Data Format

The knowledge data is stored in JSON format:

```json
{
  "knownedge_bases": [
    {
      "name": "knowledge_about_astronomy",
      "domain": "Astronomy",
      "description": "Facts about celestial bodies, space phenomena, and the universe",
      "items": [
        "The sky appears blue because Earth's atmosphere scatters shorter blue wavelengths of sunlight more than longer red wavelengths",
        "A day on Venus is longer than its year - Venus takes 243 Earth days to rotate once but only 225 Earth days to orbit the Sun"
      ]
    }
  ]
}
```

## Commands

- `:help` or `:h` - Show help
- `:clear` - Clear conversation history
- `:history` - Show conversation history
- `:info` - Show configuration information
- `:quit`, `:exit`, `:q`, `:x` - Exit the application

## Prerequisites

### Vector Database Setup
To use the real vector database functionality, you need to set up Milvus:

1. **Install Milvus**: Follow the [Milvus installation guide](https://milvus.io/docs/install_standalone-docker.md)
2. **Start Milvus**: Run `docker-compose up -d` in the Milvus directory
3. **Verify Connection**: Ensure Milvus is accessible at `localhost:19530`

### API Keys
You'll need API keys for:
- **Chat Provider**: For the main AI model (OpenAI, Anthropic, etc.)
- **Embedder Provider**: For generating embeddings (usually OpenAI)

## Troubleshooting

### Common Issues

1. **Vector Database Connection Failed**
   - Ensure Milvus is running: `docker ps | grep milvus`
   - Check endpoint in config: should be `localhost:19530`
   - Verify network connectivity

2. **Embedder Provider Error**
   - Check API key for embedder provider
   - Verify embedder model name (e.g., `text-embedding-3-small`)
   - Ensure embedder provider supports the specified model

3. **Configuration File Not Found**
   - Use `--config` flag to specify config file path
   - Ensure config file exists and is readable
   - Check JSON syntax in config file 