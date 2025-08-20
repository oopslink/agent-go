package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"

	"github.com/oopslink/agent-go/pkg/core/tools"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

const (
	KNOWLEDGE_RETRIVER_TOOL = "ag_knowledge_retriever"
)

var _ tools.Tool = &KnowledgeRetrieverToolWrapper{}

func IsKnowledgeRetrieverTool(toolName string) bool {
	return toolName == KNOWLEDGE_RETRIVER_TOOL
}

// SelectionStrategy defines how to select knowledge bases
type SelectionStrategy interface {
	Select(bases []KnowledgeBase, domains []string) ([]KnowledgeBase, error)
}

// MaxCountStrategy selects up to a maximum number of knowledge bases
type MaxCountStrategy struct {
	MaxCount int // -1 means select all
}

func NewMaxCountStrategy(maxCount int) *MaxCountStrategy {
	return &MaxCountStrategy{MaxCount: maxCount}
}

func (s *MaxCountStrategy) Select(bases []KnowledgeBase, domains []string) ([]KnowledgeBase, error) {
	if len(bases) == 0 {
		return nil, errors.Errorf(ErrorCodeNoKnowledgeBaseFound,
			"no knowledge bases available")
	}

	// If no domain filter specified, return all bases (respecting max count)
	if len(domains) == 0 {
		if s.MaxCount == -1 || len(bases) <= s.MaxCount {
			return bases, nil
		}
		return bases[:s.MaxCount], nil
	}

	// Filter bases by domains
	var selectedBases []KnowledgeBase
	for _, base := range bases {
		metadata := base.GetMetadata()
		if metadata == nil {
			continue
		}

		// Check if any of the requested domains match the base's domains
		for _, requestedDomain := range domains {
			for _, baseDomain := range metadata.Domains {
				if requestedDomain == baseDomain {
					selectedBases = append(selectedBases, base)
					break // Found a match for this base, move to next base
				}
			}
		}
	}

	if len(selectedBases) == 0 {
		return nil, errors.Errorf(ErrorCodeNoKnowledgeBaseFound,
			"no knowledge bases found for domains: %v", domains)
	}

	// Apply max count limit
	if s.MaxCount != -1 && len(selectedBases) > s.MaxCount {
		selectedBases = selectedBases[:s.MaxCount]
	}

	return selectedBases, nil
}

type KnowledgeRetrieverParams struct {
	Query          string
	MaxResults     int
	ScoreThreshold float32
	Domains        []string
	SearchOptions  []SearchOption
	MaxBases       int // Maximum number of knowledge bases to select (-1 for all)
}

func WrapperAsRetrieverTool(bases ...KnowledgeBase) *KnowledgeRetrieverToolWrapper {
	return &KnowledgeRetrieverToolWrapper{
		knowledgeBases: bases,
		strategy:       NewMaxCountStrategy(-1), // Default: select all
	}
}

func WrapperAsRetrieverToolWithStrategy(bases []KnowledgeBase, strategy SelectionStrategy) *KnowledgeRetrieverToolWrapper {
	return &KnowledgeRetrieverToolWrapper{
		knowledgeBases: bases,
		strategy:       strategy,
	}
}

type KnowledgeRetrieverToolWrapper struct {
	knowledgeBases []KnowledgeBase
	strategy       SelectionStrategy
}

func (k *KnowledgeRetrieverToolWrapper) toolName() string {
	return KNOWLEDGE_RETRIVER_TOOL
}

func (k *KnowledgeRetrieverToolWrapper) Descriptor() *llms.ToolDescriptor {
	// Collect available collections and domains from knowledge bases
	availableCollections := make(map[string]bool)
	availableDomains := make(map[string]bool)

	for _, base := range k.knowledgeBases {
		metadata := base.GetMetadata()
		if metadata == nil {
			continue
		}

		// Add domains
		for _, domain := range metadata.Domains {
			availableDomains[domain] = true
		}

		// Note: Collections are typically managed at the storage level
		// For now, we'll provide a generic description
		availableCollections["default"] = true
	}

	// Build domain description
	domainDescription := "Filter by knowledge base domains (optional)"
	if len(availableDomains) > 0 {
		domainList := make([]string, 0, len(availableDomains))
		for domain := range availableDomains {
			domainList = append(domainList, domain)
		}
		domainDescription = fmt.Sprintf("Filter by knowledge base domains. Available domains: %v", domainList)
	}

	// Build collection description
	collectionDescription := "Specific collection to search in (optional)"
	if len(availableCollections) > 0 {
		collectionList := make([]string, 0, len(availableCollections))
		for collection := range availableCollections {
			collectionList = append(collectionList, collection)
		}
		collectionDescription = fmt.Sprintf("Specific collection to search in. Available collections: %v", collectionList)
	}

	return &llms.ToolDescriptor{
		Name:        k.toolName(),
		Description: "Search and retrieve relevant knowledge from knowledge bases",
		Parameters: &llms.Schema{
			Type: "object",
			Properties: map[string]*llms.Schema{
				"query": {
					Type:        "string",
					Description: "The search query to find relevant knowledge",
				},
				"max_results": {
					Type:        "integer",
					Description: "Maximum number of results to return (default: 10)",
				},
				"score_threshold": {
					Type:        "number",
					Description: "Minimum similarity score threshold (0.0 to 1.0, default: 0.0)",
				},
				"collection": {
					Type:        "string",
					Description: collectionDescription,
				},
				"domains": {
					Type:        "array",
					Description: domainDescription,
					Items: &llms.Schema{
						Type: "string",
					},
				},
				"max_bases": {
					Type:        "integer",
					Description: "Maximum number of knowledge bases to select (-1 for all, default: -1)",
				},
			},
			Required: []string{"query"},
		},
	}
}

func (k *KnowledgeRetrieverToolWrapper) Call(ctx context.Context, toolCall *llms.ToolCall) (*llms.ToolCallResult, error) {
	retrieveParams := k.makeKnowledgeRetrieverParams(toolCall)

	// select knowledge bases to query
	bases, err := k.selectKnowledgeBases(ctx, retrieveParams)
	if err != nil {
		return nil, err
	}

	var knowledges []string
	for _, base := range bases {
		items, err := base.Search(ctx, retrieveParams.Query, retrieveParams.SearchOptions...)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			doc := item.ToDocument()
			if doc != nil {
				raw, err := json.Marshal(map[string]interface{}{
					"id":       doc.Id,
					"name":     doc.Name,
					"content":  doc.Content,
					"metadata": doc.Metadata,
				})
				if err != nil {
					return nil, err
				}
				knowledges = append(knowledges, string(raw))
			}
		}
	}

	return &llms.ToolCallResult{
		ToolCallId: toolCall.ToolCallId,
		Name:       k.toolName(),
		Result: map[string]any{
			"knowledges": knowledges,
			"count":      len(knowledges),
		},
	}, nil
}

func (k *KnowledgeRetrieverToolWrapper) selectKnowledgeBases(ctx context.Context, params *KnowledgeRetrieverParams) ([]KnowledgeBase, error) {
	// Update strategy if max_bases is specified
	if params.MaxBases != 0 {
		if maxCountStrategy, ok := k.strategy.(*MaxCountStrategy); ok {
			maxCountStrategy.MaxCount = params.MaxBases
		}
	}

	return k.strategy.Select(k.knowledgeBases, params.Domains)
}

func (k *KnowledgeRetrieverToolWrapper) makeKnowledgeRetrieverParams(toolCall *llms.ToolCall) *KnowledgeRetrieverParams {
	params := &KnowledgeRetrieverParams{
		Query:          "",
		MaxResults:     10,
		ScoreThreshold: 0.0,
		Domains:        []string{},
		MaxBases:       -1, // Default: select all
	}

	// Parse arguments from tool call
	if toolCall.Arguments != nil {
		if query, ok := toolCall.Arguments["query"].(string); ok {
			params.Query = query
		}
		if maxResults, ok := toolCall.Arguments["max_results"].(float64); ok {
			params.MaxResults = int(maxResults)
		}
		if scoreThreshold, ok := toolCall.Arguments["score_threshold"].(float64); ok {
			params.ScoreThreshold = float32(scoreThreshold)
		}
		if domains, ok := toolCall.Arguments["domains"].([]interface{}); ok {
			for _, domain := range domains {
				if domainStr, ok := domain.(string); ok {
					params.Domains = append(params.Domains, domainStr)
				}
			}
		}
		if maxBases, ok := toolCall.Arguments["max_bases"].(float64); ok {
			params.MaxBases = int(maxBases)
		}
	}

	// Build search options
	var searchOptions []SearchOption
	if params.MaxResults > 0 {
		searchOptions = append(searchOptions, WithMaxResults(params.MaxResults))
	}
	if params.ScoreThreshold > 0 {
		searchOptions = append(searchOptions, WithScoreThreshold(params.ScoreThreshold))
	}

	params.SearchOptions = searchOptions
	return params
}
