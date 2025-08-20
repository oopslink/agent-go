// Package anthropic provides Anthropic AI model provider implementations for the agent-go framework.
// This file contains the chat provider implementation for Anthropic's Claude models.
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

// init registers the Anthropic chat provider with the global provider registry.
// This function is called automatically when the package is imported.
func init() {
	_ = llms.RegisterChatProvider(ModelProviderAnthropic, newChatProvider)
}

// newChatProvider creates a new Anthropic chat provider instance.
// It configures the provider with API key, base URL, and other options.
func newChatProvider(opts ...llms.ProviderOption) (llms.ChatProvider, error) {
	options := llms.OfProviderOptions(opts...)

	var requestOptions []option.RequestOption

	// Set API key from options or environment variable
	apiKey := options.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey != "" {
		requestOptions = append(requestOptions, option.WithAPIKey(apiKey))
	}

	// Check for custom base URL from options or environment variable
	baseUrl := options.BaseUrl
	if baseUrl == "" {
		baseUrl = os.Getenv("ANTHROPIC_BASE_URL")
	}
	if baseUrl != "" {
		klog.Infof("Using custom Anthropic base URL: %s", baseUrl)
		requestOptions = append(requestOptions, option.WithBaseURL(baseUrl))
	}

	// Configure HTTP client with SSL verification settings
	httpClient := utils.CreateHTTPClient(options.SkipVerifySSL)
	requestOptions = append(requestOptions, option.WithHTTPClient(httpClient))

	if options.Debug {
		klog.Infof("[DEBUG] AnthropicProvider params: ApiKey=****, BaseUrl=%s, SkipVerifySSL=%v", baseUrl, options.SkipVerifySSL)
	}

	client := anthropic.NewClient(requestOptions...)
	return &anthropicChatProvider{
		client: &client,
		debug:  options.Debug,
	}, nil
}

// anthropicChatProvider implements the ChatProvider interface for Anthropic models.
var _ llms.ChatProvider = &anthropicChatProvider{}

type anthropicChatProvider struct {
	client *anthropic.Client
	debug  bool
}

// Close implements the io.Closer interface.
// Currently no cleanup is needed for the Anthropic client.
func (a *anthropicChatProvider) Close() error {
	return nil
}

// IsRetryableError determines if an error should trigger a retry.
// Delegates to the global provider error handling logic.
func (a *anthropicChatProvider) IsRetryableError(err error) bool {
	return errors.IsRetryableError(err)
}

// NewChat creates a new chat session with the given system prompt and model.
// Validates that the model supports completion before creating the chat.
func (a *anthropicChatProvider) NewChat(systemPrompt string, model *llms.Model) (llms.Chat, error) {
	if !model.IsSupport(llms.ModelFeatureCompletion) {
		return nil, errors.Errorf(llms.ErrorCodeModelFeatureNotMatched,
			"model %s is not support completion", model.ModelId.String())
	}
	return &anthropicChat{
		client:       a.client,
		systemPrompt: systemPrompt,
		model:        model,
		debug:        a.debug,
	}, nil
}

// anthropicChat implements the Chat interface for Anthropic models.
var _ llms.Chat = &anthropicChat{}

type anthropicChat struct {
	client       *anthropic.Client
	systemPrompt string
	model        *llms.Model
	debug        bool
}

// Send sends messages to the Anthropic API and returns an iterator for responses.
// Supports both streaming and non-streaming modes based on the provided options.
func (a *anthropicChat) Send(ctx context.Context, messages []*llms.Message, options ...llms.ChatOption) (llms.ChatResponseIterator, error) {
	opts := &llms.ChatOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Streaming {
		return a.stream(ctx, messages, opts)
	} else {
		response, err := a.sendOnce(ctx, messages, opts)
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

func (a *anthropicChat) sendOnce(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (*llms.ChatResponse, error) {
	params, err := a.makeMessageNewParams(messages, opts)
	if err != nil {
		return nil, err
	}

	if a.debug {
		j, _ := json.MarshalIndent(*params, "", "  ")
		klog.Infof("[DEBUG] chat once begin, params: %s", string(j))
	}

	result, err := utils.Retry(ctx,
		func() (r *llms.ChatResponse, err error) {
			response, err := a.client.Messages.New(ctx, *params)
			if err != nil {
				if a.isRetryable(err) {
					klog.Warning("error on chat once, will retry later", "err", err)
					return nil, err
				}
				return nil, errors.Permanent(err)
			}

			if a.debug {
				if response != nil {
					b, _ := json.MarshalIndent(response, "", "  ")
					klog.Infof("[DEBUG] Anthropic response: %s", string(b))
				} else {
					klog.Infof("[DEBUG] Anthropic response: nil, err=%v", err)
				}
			}

			finishReason, message := a.makeMessageFromAnthropicMessage(response)
			usage := a.getUsageStats(&response.Usage)

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

func (a *anthropicChat) stream(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (llms.ChatResponseIterator, error) {
	params, err := a.makeMessageNewParams(messages, opts)
	if err != nil {
		return nil, err
	}

	if a.debug {
		j, _ := json.MarshalIndent(*params, "", "  ")
		klog.Infof("[DEBUG] chat streaming begin, params: %s", string(j))
	}

	assistant := llms.MessageCreator{Role: llms.MessageRoleAssistant}

	stream := a.client.Messages.NewStreaming(ctx, *params)

	messageId := utils.GenerateUUID()
	acc := anthropic.Message{}
	return func(yield func(*llms.ChatResponse, error) bool) {
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()
			acc.Accumulate(event)

			if a.debug {
				b, _ := json.MarshalIndent(event, "", "  ")
				klog.Infof("[DEBUG] Anthropic response: %s", string(b))
			}

			switch event := event.AsAny().(type) {
			case anthropic.ContentBlockStartEvent:
				// no-ops

			case anthropic.ContentBlockDeltaEvent:
				if event.Delta.Type == "thinking_delta" && event.Delta.Thinking != "" {
					message := llms.Message{
						MessageId: messageId,
						Model:     a.model.ModelId,
						Creator:   assistant,
						Parts:     []llms.Part{llms.NewTextPartBuilder().Text(event.Delta.Thinking).Build()},
						Timestamp: time.Now(),
					}
					if !yield(&llms.ChatResponse{
						Message: message,
					}, nil) {
						return
					}
				} else if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
					message := llms.Message{
						MessageId: messageId,
						Model:     a.model.ModelId,
						Creator:   assistant,
						Parts:     []llms.Part{llms.NewTextPartBuilder().Text(event.Delta.Text).Build()},
						Timestamp: time.Now(),
					}
					if !yield(&llms.ChatResponse{
						Message: message,
					}, nil) {
						return
					}
				} else if event.Delta.Type == "input_json_delta" {
					// TODO: support tool use delta
				}
			case anthropic.ContentBlockStopEvent:
				// TODO: support tool use stop

			case anthropic.MessageStopEvent:
				// no-ops, handled by the last message
			}
		}

		err = stream.Err()
		if err == nil || errors.Is(err, io.EOF) {
			content := ""
			for _, block := range acc.Content {
				if text, ok := block.AsAny().(anthropic.TextBlock); ok {
					content += text.Text
				}
			}
		}

		if err == nil || errors.Is(err, io.EOF) {
			usage := a.getUsageStats(&acc.Usage)

			finishReason, finalMessage := a.makeMessageFromAnthropicMessage(&acc)

			var toolCallParts []llms.Part
			for idx := range finalMessage.Parts {
				part := finalMessage.Parts[idx]
				if toolCall, ok := part.(*llms.ToolCall); ok {
					toolCallParts = append(toolCallParts, toolCall)
				}
			}

			if !yield(&llms.ChatResponse{
				Message: llms.Message{
					MessageId: messageId,
					Model:     a.model.ModelId,
					Creator:   assistant,
					Parts:     toolCallParts,
					Timestamp: time.Now(),
				},
				Usage:        usage,
				FinishReason: finishReason,
			}, nil) {
				return
			}
		} else {
			klog.Warningf("error on streaming: [%s]", err)
		}
	}, nil
}

func (a *anthropicChat) makeMessageFromAnthropicMessage(response *anthropic.Message) (llms.FinishReason, llms.Message) {
	finishReason := llms.FinishReasonUnknown
	message := llms.Message{
		Creator:   llms.MessageCreator{Role: llms.MessageRoleAssistant},
		MessageId: response.ID,
		Model:     a.model.ModelId,
		Timestamp: time.Now(),
	}

	for _, block := range response.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			part := llms.NewTextPartBuilder().Text(variant.Text).Build()
			message.Parts = append(message.Parts, part)
		case anthropic.ThinkingBlock:
			part := llms.NewTextPartBuilder().Text(variant.Thinking).Build()
			message.Parts = append(message.Parts, part)
		case anthropic.RedactedThinkingBlock:
			part := llms.NewTextPartBuilder().Text(variant.Data).Build()
			message.Parts = append(message.Parts, part)
		case anthropic.ToolUseBlock:
			message.Parts = append(message.Parts, a.toToolCall(variant))
		case anthropic.ServerToolUseBlock:
			// TODO: support server tool use
		case anthropic.WebSearchToolResultBlock:
			// TODO: support web search tool result
		default:
			klog.Warningf("no variant present: %T", variant)
		}
	}

	finishReason = a.toFinishReason(response.StopReason)
	return finishReason, message
}

func (a *anthropicChat) toToolCall(variant anthropic.ToolUseBlock) *llms.ToolCall {
	var args map[string]any
	if variant.Input != nil {
		if err := json.Unmarshal(variant.Input, &args); err != nil {
			klog.InfoS("error unmarshalling function arguments", "tool", variant.Name, "err", err)
			args = make(map[string]any)
		}
	} else {
		args = make(map[string]any)
	}
	return &llms.ToolCall{
		ToolCallId: variant.ID,
		Name:       variant.Name,
		Arguments:  args,
	}
}

func (a *anthropicChat) makeMessageNewParams(
	messages []*llms.Message, opts *llms.ChatOptions) (*anthropic.MessageNewParams, error) {

	params := &anthropic.MessageNewParams{
		Model: anthropic.Model(a.model.ApiModelName),
	}

	systemInstruction := a.makeSystemInstruction(messages)
	if len(systemInstruction) > 0 {
		params.System = []anthropic.TextBlockParam{
			{
				Text: systemInstruction,
				CacheControl: anthropic.CacheControlEphemeralParam{
					Type: "ephemeral",
				},
			},
		}
	}

	anthropicMessages, err := a.convertToAnthropicMessages(messages)
	if err != nil {
		return nil, err
	}
	params.Messages = anthropicMessages

	anthropicTools, err := a.convertToAnthropicTools(opts.Tools)
	if err != nil {
		return nil, err
	}
	if len(anthropicTools) > 0 {
		params.Tools = anthropicTools
	}

	if opts.Temperature != nil {
		params.Temperature = anthropic.Float(*opts.Temperature)
	}

	if opts.TopP != nil {
		params.TopP = anthropic.Float(*opts.TopP)
	}

	if opts.MaxCompletionTokens != nil {
		params.MaxTokens = *opts.MaxCompletionTokens
	} else {
		params.MaxTokens = a.model.DefaultMaxTokens
	}

	if a.model.IsSupport(llms.ModelFeatureReasoning) && a.shouldThink(messages, opts) {
		params.Thinking = anthropic.ThinkingConfigParamOfEnabled(int64(float64(params.MaxTokens) * 0.8))
	}

	return params, nil
}

func (a *anthropicChat) convertToAnthropicMessages(messages []*llms.Message) ([]anthropic.MessageParam, error) {
	anthropicMessages := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		// Skip system messages as they are handled separately in the system prompt
		if msg.Creator.Role == llms.MessageRoleSystem {
			continue
		}

		// Convert message parts to content blocks
		var contentBlocks []anthropic.ContentBlockParamUnion
		for _, part := range msg.Parts {
			switch part.Type() {
			case llms.PartTypeText:
				if textPart, ok := part.(*llms.TextPart); ok {
					contentBlocks = append(contentBlocks, anthropic.NewTextBlock(textPart.Text))
				}
			case llms.PartTypeData:
				if dataPart, ok := part.(*llms.DataPart); ok {
					// Convert data to text representation
					contentBlocks = append(contentBlocks, anthropic.NewTextBlock(dataPart.MarshalJson()))
				}
			case llms.PartTypeBinary:
				if binaryPart, ok := part.(*llms.BinaryPart); ok {
					if _, isImage := llms.IsImagePart(binaryPart); isImage {
						if len(binaryPart.Content) > 0 {
							// Use the NewImageBlockBase64 helper function
							contentBlocks = append(contentBlocks, anthropic.NewImageBlockBase64(binaryPart.MIMEType, binaryPart.MarshalBase64()))
						}
					} else if _, isAudio := llms.IsAudioPart(binaryPart); isAudio {
						// no-ops, anthropic doesn't support audio
					} else {
						if len(binaryPart.Content) > 0 {
							if llms.IsPlainTextPart(binaryPart) {
								contentBlocks = append(contentBlocks,
									anthropic.NewDocumentBlock(
										anthropic.PlainTextSourceParam{
											Data: string(binaryPart.Content),
										}))
							} else if llms.IsPDFPart(binaryPart) {
								base64Content := binaryPart.MarshalBase64()
								contentBlocks = append(contentBlocks,
									anthropic.NewDocumentBlock(
										anthropic.Base64PDFSourceParam{
											Data: base64Content,
										}))
							} else {
								// TODO: support other file types
							}
						} else {
							if llms.IsPDFPart(binaryPart) {
								if binaryPart.URL != nil {
									contentBlocks = append(contentBlocks,
										anthropic.NewDocumentBlock(
											anthropic.URLPDFSourceParam{
												URL: *binaryPart.URL,
											}))
								}
							} else {
								fileName := ""
								if binaryPart.Name != nil {
									fileName = *binaryPart.Name
								}
								fileId := ""
								if binaryPart.URL != nil {
									fileId = *binaryPart.URL
								}
								lines := []string{"> reference file:"}
								if len(fileName) > 0 {
									lines = append(lines, fmt.Sprintf("> - file name: %s", fileName))
								}
								if len(fileId) > 0 {
									lines = append(lines, fmt.Sprintf("> - url: %s", fileId))
								}
								if len(lines) > 0 {
									contentBlocks = append(contentBlocks, anthropic.NewTextBlock(strings.Join(lines, "\n")))
								}
							}
						}
					}
				}
			case llms.PartTypeToolCall:
				if toolCallPart, ok := part.(*llms.ToolCall); ok {
					contentBlocks = append(contentBlocks, anthropic.ContentBlockParamUnion{
						OfToolUse: &anthropic.ToolUseBlockParam{
							ID:    toolCallPart.ToolCallId,
							Name:  toolCallPart.Name,
							Input: toolCallPart.Arguments,
						},
					})
				}
			case llms.PartTypeToolCallResult:
				if toolResultPart, ok := part.(*llms.ToolCallResult); ok {
					contentBlocks = append(contentBlocks, anthropic.ContentBlockParamUnion{
						OfToolResult: &anthropic.ToolResultBlockParam{
							ToolUseID: toolResultPart.ToolCallId,
							Content: []anthropic.ToolResultBlockParamContentUnion{
								{
									OfText: &anthropic.TextBlockParam{
										Text: toolResultPart.MarshalJson(),
									},
								},
							},
						},
					})
				}
			}
		}

		// Create message based on role
		var anthropicMsg anthropic.MessageParam
		switch msg.Creator.Role {
		case llms.MessageRoleUser:
			anthropicMsg = anthropic.NewUserMessage(contentBlocks...)
		case llms.MessageRoleAssistant:
			anthropicMsg = anthropic.NewAssistantMessage(contentBlocks...)
		case llms.MessageRoleTool:
			// Anthropic doesn't have separate tool role, treat as user
			anthropicMsg = anthropic.NewUserMessage(contentBlocks...)
		default:
			// Default to user message
			anthropicMsg = anthropic.NewUserMessage(contentBlocks...)
		}

		anthropicMessages = append(anthropicMessages, anthropicMsg)
	}

	return anthropicMessages, nil
}

func (a *anthropicChat) convertToAnthropicTools(tools []*llms.ToolDescriptor) ([]anthropic.ToolUnionParam, error) {
	if len(tools) == 0 {
		return []anthropic.ToolUnionParam{}, nil
	}

	anthropicTools := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		// Convert the tool's parameter schema to a simple map format
		var inputSchema anthropic.ToolInputSchemaParam

		if tool.Parameters != nil {
			// Convert schema to map format for Properties
			properties := make(map[string]any)
			if tool.Parameters.Properties != nil {
				for key, prop := range tool.Parameters.Properties {
					propMap := make(map[string]any)
					propMap["type"] = string(prop.Type)
					if prop.Description != "" {
						propMap["description"] = prop.Description
					}
					properties[key] = propMap
				}
			}
			inputSchema.Properties = properties

			if len(tool.Parameters.Required) > 0 {
				inputSchema.Required = tool.Parameters.Required
			}
		}

		// Create tool param using the OfTool field
		anthropicTool := anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: inputSchema,
				Type:        anthropic.ToolTypeCustom,
			},
		}
		anthropicTools = append(anthropicTools, anthropicTool)
	}

	return anthropicTools, nil
}

func (a *anthropicChat) toFinishReason(reason anthropic.StopReason) llms.FinishReason {
	if len(reason) == 0 {
		return ""
	}
	switch reason {
	case anthropic.StopReasonEndTurn:
		return llms.FinishReasonNormalEnd
	case anthropic.StopReasonMaxTokens:
		return llms.FinishReasonMaxTokens
	case anthropic.StopReasonStopSequence:
		return llms.FinishReasonNormalEnd
	case anthropic.StopReasonToolUse:
		return llms.FinishReasonToolUse
	default:
		return llms.FinishReasonUnknown
	}
}

func (a *anthropicChat) getUsageStats(usage *anthropic.Usage) llms.UsageMetadata {
	return llms.UsageMetadata{
		InputTokens:         usage.InputTokens,
		OutputTokens:        usage.OutputTokens,
		CacheCreationTokens: usage.CacheCreationInputTokens,
		CacheReadTokens:     usage.CacheReadInputTokens,
	}
}

func (a *anthropicChat) isRetryable(err error) bool {
	return errors.IsRetryableError(err)
}

func (a *anthropicChat) shouldThink(messages []*llms.Message, opts *llms.ChatOptions) bool {
	// TODO check if need thinking
	return false
}

func (a *anthropicChat) makeSystemInstruction(messages []*llms.Message) string {
	return llms.MakeSystemInstruction(a.systemPrompt, messages)
}
