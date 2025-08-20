package gemini

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/genai"
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/journal"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func init() {
	_ = llms.RegisterChatProvider(ModelProviderGemini, newChatProvider)
}

func newChatProvider(opts ...llms.ProviderOption) (llms.ChatProvider, error) {
	options := llms.OfProviderOptions(opts...) // Create client config

	config := &genai.ClientConfig{
		Project: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		Backend: genai.BackendGeminiAPI,
	}

	// Set API key
	apiKey := options.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey != "" {
		config.APIKey = apiKey
	}

	// Create client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, errors.Errorf(llms.ErrorCodeCreateChatProviderFailed,
			"failed to create Gemini client: %s", err.Error())
	}

	if options.Debug {
		journal.Info("llm", "gemini", "provider initialized",
			"skipVerifySSL", options.SkipVerifySSL)
	}

	return &geminiChatProvider{
		client: client,
		debug:  options.Debug,
	}, nil
}

var _ llms.ChatProvider = &geminiChatProvider{}

type geminiChatProvider struct {
	client *genai.Client
	debug  bool
}

func (g *geminiChatProvider) Close() error {
	// New API doesn't have Close method
	return nil
}

func (g *geminiChatProvider) IsRetryableError(err error) bool {
	return errors.IsRetryableError(err)
}

func (g *geminiChatProvider) NewChat(systemPrompt string, model *llms.Model) (llms.Chat, error) {
	if !model.IsSupport(llms.ModelFeatureCompletion) {
		return nil, errors.Errorf(llms.ErrorCodeModelFeatureNotMatched,
			"model %s is not support completion", model.ModelId.String())
	}
	return &geminiChat{
		client:       g.client,
		systemPrompt: systemPrompt,
		model:        model,
		debug:        g.debug,
	}, nil
}

var _ llms.Chat = &geminiChat{}

type geminiChat struct {
	client       *genai.Client
	systemPrompt string
	model        *llms.Model
	debug        bool
}

func (g *geminiChat) Send(ctx context.Context, messages []*llms.Message, options ...llms.ChatOption) (llms.ChatResponseIterator, error) {
	opts := &llms.ChatOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Streaming {
		return g.stream(ctx, messages, opts)
	} else {
		response, err := g.sendOnce(ctx, messages, opts)
		if err != nil {
			return nil, err
		}
		onceValue := utils.OfOnceValue(response)
		return func(yield func(*llms.ChatResponse, error) bool) {
			onceValue.Get(func(v *llms.ChatResponse) {
				yield(response, nil)
			})
		}, nil
	}
}

func (g *geminiChat) sendOnce(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (*llms.ChatResponse, error) {
	// Create generation config
	systemInstruction := g.makeSystemInstruction(messages)
	config := g.createGenerationConfig(systemInstruction, opts)

	// Convert messages to Gemini format
	history, currentParts, err := g.convertMessages(messages)
	if err != nil {
		return nil, err
	}

	// Create chat session
	chat, err := g.client.Chats.Create(ctx, g.model.ApiModelName, config, history)
	if err != nil {
		return nil, errors.Errorf(llms.ErrorCodeChatSessionFailed,
			"failed to create chat session: %s", err.Error())
	}

	if g.debug {
		journal.Info("llm", "gemini", "chat once begin",
			"modelId", g.model.ModelId, "historyLength", len(history), "config", config, "history", history)
	}

	result, err := utils.Retry(ctx,
		func() (r *llms.ChatResponse, err error) {
			// Send message
			response, err := chat.SendMessage(ctx, currentParts...)
			if err != nil {
				if g.isRetryable(err) {
					klog.Warning("error on chat once, will retry later", "err", err)
					return nil, err
				}
				return nil, errors.Permanent(err)
			}

			if g.debug {
				if response != nil {
					journal.Info("llm", "gemini", "chat once response", "response", response)
				} else {
					journal.Info("llm", "gemini", "chat once response", "err", err)
				}
			}

			finishReason, message := g.makeMessageFromResponse(utils.GenerateUUID(), response)
			usage := g.getUsageStats(response)
			journal.AccumulateUsage("chat", usage.AsMap())

			return &llms.ChatResponse{
				Message:      message,
				Usage:        usage,
				FinishReason: finishReason,
			}, nil
		},
		utils.WithBackOff(utils.NewExponentialBackOff()),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *geminiChat) stream(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (llms.ChatResponseIterator, error) {
	// Create generation config
	systemInstruction := g.makeSystemInstruction(messages)
	config := g.createGenerationConfig(systemInstruction, opts)

	// Convert messages to Gemini format
	history, currentParts, err := g.convertMessages(messages)
	if err != nil {
		return nil, err
	}

	// Create chat session
	chat, err := g.client.Chats.Create(ctx, g.model.ApiModelName, config, history)
	if err != nil {
		return nil, errors.Errorf(llms.ErrorCodeChatSessionFailed,
			"failed to create chat session: %s", err.Error())
	}

	if g.debug {
		journal.Info("llm", "gemini", "chat streaming begin",
			"modelId", g.model.ModelId, "historyLength", len(history), "config", config, "history", history)
	}

	assistant := llms.MessageCreator{Role: llms.MessageRoleAssistant}

	// Send message with streaming
	iter := chat.SendMessageStream(ctx, currentParts...)

	messageId := utils.GenerateUUID()
	acc := newGeminiChatCompletionAccumulator()

	return func(yield func(*llms.ChatResponse, error) bool) {

		for response, err := range iter {
			if err != nil {
				if errors.Is(err, io.EOF) {
					// End of stream
					break
				}
				if !yield(nil, err) {
					return
				}
				continue
			}

			acc.AddChunk(response)

			if g.debug {
				journal.Info("llm", "gemini", "streaming response", "response", response)
			}

			// Process candidates
			if len(response.Candidates) > 0 {
				finishReason, message := g.makeMessageFromResponse(messageId, response)
				if !yield(&llms.ChatResponse{
					Message:      message,
					FinishReason: finishReason,
				}, nil) {
					return
				}
			}
		}

		usage := g.getUsageStats(&acc.GenerateContentResponse)
		journal.AccumulateUsage("chat", usage.AsMap())

		finishReason, _ := g.makeMessageFromResponse(messageId, &acc.GenerateContentResponse)

		if !yield(&llms.ChatResponse{
			Message: llms.Message{
				MessageId: messageId,
				Model:     g.model.ModelId,
				Creator:   assistant,
				Timestamp: time.Now(),
			},
			Usage:        usage,
			FinishReason: finishReason,
		}, nil) {
			return
		}
	}, nil
}

func (g *geminiChat) createGenerationConfig(systemInstruction string, opts *llms.ChatOptions) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}

	// Set system instruction
	if len(systemInstruction) > 0 {
		systemContent := genai.Text(systemInstruction)
		if len(systemContent) > 0 {
			config.SystemInstruction = systemContent[0]
		}
	}

	// Set generation config
	if opts.Temperature != nil {
		config.Temperature = genai.Ptr(float32(*opts.Temperature))
	}

	if opts.TopP != nil {
		config.TopP = genai.Ptr(float32(*opts.TopP))
	}

	if opts.MaxCompletionTokens != nil {
		config.MaxOutputTokens = int32(*opts.MaxCompletionTokens)
	}

	// Configure tools if provided
	if len(opts.Tools) > 0 {
		tools, err := g.convertToGeminiTools(opts.Tools)
		if err != nil {
			journal.Warning("llm", "gemini", "Failed to convert tools for Gemini", "error", err)
		} else {
			config.Tools = tools
		}
	}

	return config
}

func (g *geminiChat) convertMessages(messages []*llms.Message) ([]*genai.Content, []genai.Part, error) {
	var history []*genai.Content
	var currentParts []genai.Part

	for i, msg := range messages {
		content := &genai.Content{}

		// Set role
		switch msg.Creator.Role {
		case llms.MessageRoleUser:
			content.Role = genai.RoleUser
		case llms.MessageRoleAssistant:
			content.Role = genai.RoleModel
		case llms.MessageRoleTool:
			content.Role = genai.RoleUser // Gemini doesn't have separate tool role
		default:
			content.Role = genai.RoleUser
		}

		// Convert parts
		for _, part := range msg.Parts {
			switch part.Type() {
			case llms.PartTypeText:
				if textPart, ok := part.(*llms.TextPart); ok {
					content.Parts = append(content.Parts, genai.NewPartFromText(textPart.Text))
				}
			case llms.PartTypeData:
				if dataPart, ok := part.(*llms.DataPart); ok {
					content.Parts = append(content.Parts, genai.NewPartFromText(fmt.Sprintf("```\n%s\n```", dataPart.MarshalJson())))
				}
			case llms.PartTypeBinary:
				if binaryPart, ok := part.(*llms.BinaryPart); ok {
					if _, isImage := llms.IsImagePart(binaryPart); isImage {
						if len(binaryPart.Content) > 0 {
							content.Parts = append(content.Parts, genai.NewPartFromBytes(binaryPart.Content, binaryPart.MIMEType))
						}
					} else {
						// For non-image files, convert to text
						fileName := ""
						if binaryPart.Name != nil {
							fileName = *binaryPart.Name
						}
						content.Parts = append(content.Parts, genai.NewPartFromText(fmt.Sprintf("File: %s", fileName)))
					}
				}
			case llms.PartTypeToolCall:
				if toolCallPart, ok := part.(*llms.ToolCall); ok {
					// Convert tool call to function call
					part := genai.NewPartFromFunctionCall(toolCallPart.Name, toolCallPart.Arguments)
					part.FunctionCall.ID = toolCallPart.ToolCallId // Use original ToolCallId
					content.Parts = append(content.Parts, part)
				}
			case llms.PartTypeToolCallResult:
				if toolResultPart, ok := part.(*llms.ToolCallResult); ok {
					// Convert tool result to function response
					part := genai.NewPartFromFunctionResponse(
						toolResultPart.Name,
						map[string]interface{}{
							"result": toolResultPart.MarshalJson(),
						})
					part.FunctionResponse.ID = toolResultPart.ToolCallId // Use original ToolCallId
					content.Parts = append(content.Parts, part)
				}
			}
		}

		// If this is the last message, use it as current message
		if i == len(messages)-1 {
			currentParts = make([]genai.Part, len(content.Parts))
			for j, part := range content.Parts {
				currentParts[j] = *part
			}
		} else {
			history = append(history, content)
		}
	}

	return history, currentParts, nil
}

func (g *geminiChat) convertToGeminiTools(tools []*llms.ToolDescriptor) ([]*genai.Tool, error) {
	var geminiTools []*genai.Tool

	for _, tool := range tools {
		// Convert tool descriptor to Gemini function declaration
		funcDecl := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
		}

		// Convert parameters schema
		if tool.Parameters != nil {
			schema := g.convertSchemaToGenai(tool.Parameters)
			funcDecl.Parameters = schema
		}

		geminiTool := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		}

		geminiTools = append(geminiTools, geminiTool)
	}

	return geminiTools, nil
}

func (g *geminiChat) convertSchemaToGenai(schema *llms.Schema) *genai.Schema {
	if schema == nil {
		return nil
	}

	genaiSchema := &genai.Schema{
		Type:        g.convertSchemaType(schema.Type),
		Description: schema.Description,
	}

	if len(schema.Required) > 0 {
		genaiSchema.Required = schema.Required
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]*genai.Schema)
		for key, prop := range schema.Properties {
			props[key] = g.convertSchemaToGenai(prop)
		}
		genaiSchema.Properties = props
	}

	if schema.Items != nil {
		genaiSchema.Items = g.convertSchemaToGenai(schema.Items)
	}

	return genaiSchema
}

func (g *geminiChat) convertSchemaType(schemaType llms.SchemaType) genai.Type {
	switch schemaType {
	case llms.TypeString:
		return genai.TypeString
	case llms.TypeNumber:
		return genai.TypeNumber
	case llms.TypeInteger:
		return genai.TypeInteger
	case llms.TypeBoolean:
		return genai.TypeBoolean
	case llms.TypeArray:
		return genai.TypeArray
	case llms.TypeObject:
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}

func (g *geminiChat) makeMessageFromResponse(
	messageId string, response *genai.GenerateContentResponse) (llms.FinishReason, llms.Message) {
	finishReason := llms.FinishReasonUnknown
	message := llms.Message{
		Creator:   llms.MessageCreator{Role: llms.MessageRoleAssistant},
		MessageId: messageId,
		Model:     g.model.ModelId,
		Timestamp: time.Now(),
	}

	if len(response.Candidates) > 0 {
		candidate := response.Candidates[0]

		// Convert finish reason
		finishReason = g.toFinishReason(candidate.FinishReason)

		// Convert content
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					textPart := llms.NewTextPartBuilder().Text(part.Text).Build()
					message.Parts = append(message.Parts, textPart)
				}
				if part.FunctionCall != nil {
					toolCall := &llms.ToolCall{
						ToolCallId: part.FunctionCall.ID,
						Name:       part.FunctionCall.Name,
						Arguments:  part.FunctionCall.Args,
					}
					message.Parts = append(message.Parts, toolCall)
					finishReason = llms.FinishReasonToolUse
				}
			}
		}
	}

	return finishReason, message
}

func (g *geminiChat) toFinishReason(reason genai.FinishReason) llms.FinishReason {
	if len(reason) == 0 {
		return ""
	}
	switch reason {
	case genai.FinishReasonStop:
		return llms.FinishReasonNormalEnd
	case genai.FinishReasonMaxTokens:
		return llms.FinishReasonMaxTokens
	case genai.FinishReasonSafety:
		return llms.FinishReasonDenied
	case genai.FinishReasonRecitation:
		return llms.FinishReasonDenied
	case genai.FinishReasonOther:
		return llms.FinishReasonUnknown
	default:
		return llms.FinishReasonUnknown
	}
}

func (g *geminiChat) getUsageStats(response *genai.GenerateContentResponse) llms.UsageMetadata {
	if response == nil || response.UsageMetadata == nil {
		return llms.UsageMetadata{}
	}

	usage := response.UsageMetadata
	return llms.UsageMetadata{
		InputTokens:         int64(usage.PromptTokenCount),
		OutputTokens:        int64(usage.CandidatesTokenCount),
		CacheReadTokens:     int64(usage.CachedContentTokenCount),
		CacheCreationTokens: 0, // Gemini doesn't provide this
	}
}

func (g *geminiChat) isRetryable(err error) bool {
	return errors.IsRetryableError(err)
}

func (g *geminiChat) makeSystemInstruction(messages []*llms.Message) string {
	return llms.MakeSystemInstruction(g.systemPrompt, messages)
}

func newGeminiChatCompletionAccumulator() *geminiChatCompletionAccumulator {
	return &geminiChatCompletionAccumulator{}
}

type geminiChatCompletionAccumulator struct {
	genai.GenerateContentResponse
}

func (a *geminiChatCompletionAccumulator) AddChunk(response *genai.GenerateContentResponse) {
	if response == nil {
		return
	}

	// Initialize if this is the first chunk
	if a.Candidates == nil {
		a.Candidates = make([]*genai.Candidate, 0)
	}

	// Process each candidate in the response
	for i, candidate := range response.Candidates {
		// Ensure we have enough candidates in accumulator
		for len(a.Candidates) <= i {
			a.Candidates = append(a.Candidates, &genai.Candidate{
				Content: &genai.Content{
					Parts: make([]*genai.Part, 0),
				},
			})
		}

		accCandidate := a.Candidates[i]

		// Update finish reason (last one wins)
		if candidate.FinishReason != "" {
			accCandidate.FinishReason = candidate.FinishReason
		}

		// Update safety ratings (last one wins)
		if len(candidate.SafetyRatings) > 0 {
			accCandidate.SafetyRatings = candidate.SafetyRatings
		}

		// Update citation metadata (last one wins)
		if candidate.CitationMetadata != nil {
			accCandidate.CitationMetadata = candidate.CitationMetadata
		}

		// Update token count (last one wins)
		if candidate.TokenCount != 0 {
			accCandidate.TokenCount = candidate.TokenCount
		}

		// Initialize content if needed
		if accCandidate.Content == nil {
			accCandidate.Content = &genai.Content{
				Parts: make([]*genai.Part, 0),
			}
		}

		// Set role from the first non-empty role
		if candidate.Content != nil && candidate.Content.Role != "" && accCandidate.Content.Role == "" {
			accCandidate.Content.Role = candidate.Content.Role
		}

		// Accumulate parts
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part == nil {
					continue
				}

				// Handle text parts - concatenate them
				if part.Text != "" {
					// Find the last text part in accumulated parts or create a new one
					var lastTextPart *genai.Part
					if len(accCandidate.Content.Parts) > 0 {
						lastPart := accCandidate.Content.Parts[len(accCandidate.Content.Parts)-1]
						if lastPart.Text != "" {
							lastTextPart = lastPart
						}
					}

					if lastTextPart != nil {
						// Concatenate to existing text part
						lastTextPart.Text += part.Text
					} else {
						// Create new text part
						newPart := &genai.Part{
							Text: part.Text,
						}
						accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
					}
				}

				// Handle function calls - add them as new parts
				if part.FunctionCall != nil {
					newPart := &genai.Part{
						FunctionCall: &genai.FunctionCall{
							ID:   part.FunctionCall.ID,
							Name: part.FunctionCall.Name,
							Args: part.FunctionCall.Args,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle function responses - add them as new parts
				if part.FunctionResponse != nil {
					newPart := &genai.Part{
						FunctionResponse: &genai.FunctionResponse{
							ID:       part.FunctionResponse.ID,
							Name:     part.FunctionResponse.Name,
							Response: part.FunctionResponse.Response,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle inline data - add as new parts
				if part.InlineData != nil {
					newPart := &genai.Part{
						InlineData: &genai.Blob{
							MIMEType: part.InlineData.MIMEType,
							Data:     part.InlineData.Data,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle file data - add as new parts
				if part.FileData != nil {
					newPart := &genai.Part{
						FileData: &genai.FileData{
							MIMEType: part.FileData.MIMEType,
							FileURI:  part.FileData.FileURI,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle executable code - add as new parts
				if part.ExecutableCode != nil {
					newPart := &genai.Part{
						ExecutableCode: &genai.ExecutableCode{
							Language: part.ExecutableCode.Language,
							Code:     part.ExecutableCode.Code,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle code execution result - add as new parts
				if part.CodeExecutionResult != nil {
					newPart := &genai.Part{
						CodeExecutionResult: &genai.CodeExecutionResult{
							Outcome: part.CodeExecutionResult.Outcome,
							Output:  part.CodeExecutionResult.Output,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle thought - add as new parts
				if part.Thought {
					newPart := &genai.Part{
						Thought: true,
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}

				// Handle video metadata - add as new parts
				if part.VideoMetadata != nil {
					newPart := &genai.Part{
						VideoMetadata: &genai.VideoMetadata{
							StartOffset: part.VideoMetadata.StartOffset,
							EndOffset:   part.VideoMetadata.EndOffset,
						},
					}
					accCandidate.Content.Parts = append(accCandidate.Content.Parts, newPart)
				}
			}
		}
	}

	// Update usage metadata (accumulate token counts)
	if response.UsageMetadata != nil {
		if a.UsageMetadata == nil {
			a.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{}
		}

		// Accumulate token counts for streaming responses
		// Note: For streaming, we typically get incremental counts in each chunk
		// but some implementations might send total counts. We'll treat them as incremental.
		if a.UsageMetadata.PromptTokenCount == 0 {
			// First chunk - set prompt token count (usually only in first chunk)
			a.UsageMetadata.PromptTokenCount = response.UsageMetadata.PromptTokenCount
		}

		// Accumulate candidate tokens (incremental)
		a.UsageMetadata.CandidatesTokenCount += response.UsageMetadata.CandidatesTokenCount

		// Update total token count (recalculate from accumulated values)
		a.UsageMetadata.TotalTokenCount = a.UsageMetadata.PromptTokenCount + a.UsageMetadata.CandidatesTokenCount

		// Handle cached content tokens (usually only in first chunk)
		if response.UsageMetadata.CachedContentTokenCount > 0 {
			a.UsageMetadata.CachedContentTokenCount = response.UsageMetadata.CachedContentTokenCount
		}
	}

	// Update prompt feedback (last one wins)
	if response.PromptFeedback != nil {
		a.PromptFeedback = response.PromptFeedback
	}
}
