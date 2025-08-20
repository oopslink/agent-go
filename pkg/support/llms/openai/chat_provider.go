package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"k8s.io/klog/v2"

	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/support/llms"
)

func init() {
	_ = llms.RegisterChatProvider(ModelProviderOpenAI, newChatProvider)
}

func newChatProvider(opts ...llms.ProviderOption) (llms.ChatProvider, error) {
	options := llms.OfProviderOptions(opts...) // Create client config
	client, err := createOpenAIClient(options)
	if err != nil {
		return nil, err
	}
	return &openAIChatProvider{
		client: client,
		debug:  options.Debug,
	}, nil
}

func createOpenAIClient(options *llms.ProviderOptions) (openai.Client, error) {
	var requestOptions []option.RequestOption

	// Set options for client creation
	apiKey := options.ApiKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey != "" {
		requestOptions = append(requestOptions, option.WithAPIKey(apiKey))
	}

	// Check for custom endpoint or API base URL
	baseUrl := options.BaseUrl
	if baseUrl == "" {
		baseUrl = os.Getenv("OPENAI_BASE_URL")
	}
	if baseUrl != "" {
		klog.Infof("Using custom OpenAI base URL: %s", baseUrl)
		requestOptions = append(requestOptions, option.WithBaseURL(baseUrl))
	}

	httpClient := utils.CreateHTTPClient(options.SkipVerifySSL)
	requestOptions = append(requestOptions, option.WithHTTPClient(httpClient))

	if options.Debug {
		klog.Infof("[DEBUG] OpenAIProvider params: ApiKey=****, BaseUrl=%s, SkipVerifySSL=%v", baseUrl, options.SkipVerifySSL)
	}

	client := openai.NewClient(requestOptions...)
	return client, nil
}

var _ llms.ChatProvider = &openAIChatProvider{}

type openAIChatProvider struct {
	client openai.Client
	debug  bool
}

func (o *openAIChatProvider) Close() error {
	// No-ops
	return nil
}

func (o *openAIChatProvider) IsRetryableError(err error) bool {
	return errors.IsRetryableError(err)
}

func (o *openAIChatProvider) NewChat(systemPrompt string, model *llms.Model) (llms.Chat, error) {
	if !model.IsSupport(llms.ModelFeatureCompletion) {
		return nil, errors.Errorf(llms.ErrorCodeModelFeatureNotMatched,
			"model %s is not support completion", model.ModelId.String())
	}
	return &openAIChat{
		client: o.client,

		systemPrompt: systemPrompt,
		model:        model,
		debug:        o.debug,
	}, nil
}

var _ llms.Chat = &openAIChat{}

type openAIChat struct {
	client openai.Client

	systemPrompt string
	model        *llms.Model
	debug        bool
}

func (o *openAIChat) Send(ctx context.Context, messages []*llms.Message, options ...llms.ChatOption) (llms.ChatResponseIterator, error) {
	opts := &llms.ChatOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.Streaming {
		return o.stream(ctx, messages, opts)
	} else {
		response, err := o.sendOnce(ctx, messages, opts)
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

func (o *openAIChat) sendOnce(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (*llms.ChatResponse, error) {
	params, err := o.makeChatCompletionParams(messages, opts)
	if err != nil {
		return nil, err
	}

	if o.debug {
		j, _ := json.MarshalIndent(*params, "", "  ")
		klog.Infof("[DEBUG] chat once begin, params: %s", string(j))
	}

	result, err := utils.Retry(ctx,
		func() (r *llms.ChatResponse, err error) {

			response, err := o.client.Chat.Completions.New(ctx, *params)
			if err != nil {
				if o.isRetryable(err) {
					klog.Warning("error on chat once, will retry later", "err", err)
					return nil, err
				}
				return nil, errors.Permanent(err)
			}

			if o.debug {
				if response != nil {
					b, _ := json.MarshalIndent(response, "", "  ")
					klog.Infof("[DEBUG] OpenAI response: %s", string(b))
				} else {
					klog.Infof("[DEBUG] OpenAI response: nil, err=%v", err)
				}
			}

			finishReason, message := o.makeMessageFromChatCompletion(response)
			usage := o.getUsageStats(&response.Usage)

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

func (o *openAIChat) makeMessageFromChatCompletion(response *openai.ChatCompletion) (llms.FinishReason, llms.Message) {
	finishReason := llms.FinishReasonUnknown
	message := llms.Message{
		Creator:   llms.MessageCreator{Role: llms.MessageRoleAssistant},
		MessageId: response.ID,
		Model:     o.model.ModelId,
		Timestamp: time.Now(),
	}

	if len(response.Choices) > 0 {
		choice := response.Choices[0]
		if len(choice.Message.Content) > 0 {
			part := llms.NewTextPartBuilder().Text(choice.Message.Content).Build()
			message.Parts = append(message.Parts, part)
		}
		if toolCalls := o.toToolCalls(choice.Message.ToolCalls); len(toolCalls) > 0 {
			for idx := range toolCalls {
				toolCall := toolCalls[idx]
				message.Parts = append(message.Parts, toolCall)
			}
			finishReason = llms.FinishReasonToolUse
		} else {
			finishReason = o.toFinishReason(choice.FinishReason)
		}
	}
	return finishReason, message
}

func (o *openAIChat) stream(ctx context.Context, messages []*llms.Message, opts *llms.ChatOptions) (llms.ChatResponseIterator, error) {
	params, err := o.makeChatCompletionParams(messages, opts)
	if err != nil {
		return nil, err
	}

	if o.debug {
		j, _ := json.MarshalIndent(*params, "", "  ")
		klog.Infof("[DEBUG] chat streaming begin, params: %s", string(j))
	}

	assistant := llms.MessageCreator{Role: llms.MessageRoleAssistant}

	stream := o.client.Chat.Completions.NewStreaming(ctx, *params)

	acc := openai.ChatCompletionAccumulator{}
	return func(yield func(*llms.ChatResponse, error) bool) {
		defer stream.Close()

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if o.debug {
				b, _ := json.MarshalIndent(chunk, "", "  ")
				klog.Infof("[DEBUG] OpenAI response: %s", string(b))
			}

			// handle refusal completion
			if refusal, ok := acc.JustFinishedRefusal(); ok {
				message := llms.Message{
					MessageId: chunk.ID,
					Model:     o.model.ModelId,
					Creator:   assistant,
					Timestamp: time.Now(),
				}
				if len(refusal) > 0 {
					part := llms.NewTextPartBuilder().Text(refusal).Build()
					message.Parts = append(message.Parts, part)
				}
				yield(&llms.ChatResponse{
					Message:      message,
					FinishReason: llms.FinishReasonDenied,
				}, nil)
				return
			}

			// handle tool call completion
			if tool, ok := acc.JustFinishedToolCall(); ok {
				message := llms.Message{
					MessageId: chunk.ID,
					Model:     o.model.ModelId,
					Creator:   assistant,
					Timestamp: time.Now(),
				}
				message.Parts = append(message.Parts,
					o.toToolCall(tool.ID, tool.ChatCompletionMessageToolCallFunction))
				if !yield(&llms.ChatResponse{
					Message:      message,
					FinishReason: llms.FinishReasonToolUse,
				}, nil) {
					return
				}
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if len(delta.Content) > 0 {
					message := llms.Message{
						MessageId: chunk.ID,
						Model:     o.model.ModelId,
						Creator:   assistant,
						Parts:     []llms.Part{llms.NewTextPartBuilder().Text(delta.Content).Build()},
						Timestamp: time.Now(),
					}
					if !yield(&llms.ChatResponse{
						Message: message,
					}, nil) {
						return
					}
				}
			}
		}

		err = stream.Err()
		if err == nil || errors.Is(err, io.EOF) {
			usage := o.getUsageStats(&acc.Usage)

			finishReason, _ := o.makeMessageFromChatCompletion(&acc.ChatCompletion)

			if !yield(&llms.ChatResponse{
				Message: llms.Message{
					MessageId: acc.ChatCompletion.ID,
					Model:     o.model.ModelId,
					Creator:   assistant,
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

func (o *openAIChat) convertToOpenAIMessages(messages []*llms.Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	openaiMessages := []openai.ChatCompletionMessageParamUnion{
		o.makeSystemMessage(messages),
	}

	for _, msg := range messages {
		switch msg.Creator.Role {
		case llms.MessageRoleUser:
			content := o.makeChatCompletionContent(msg)
			userMessage := openai.UserMessage(content)
			if userMessage.OfUser != nil {
				if msg.Creator.Name != nil {
					userMessage.OfUser.Name = openai.String(*msg.Creator.Name)
				}
				openaiMessages = append(openaiMessages, userMessage)
			}
		case llms.MessageRoleAssistant:
			assistant := o.makeChatCompletionAssistantMessageParam(msg)
			if assistant != nil {
				assistantMessage := openai.ChatCompletionMessageParamUnion{OfAssistant: assistant}
				if msg.Creator.Name != nil {
					assistantMessage.OfAssistant.Name = openai.String(*msg.Creator.Name)
				}
				openaiMessages = append(openaiMessages, assistantMessage)
			}

		case llms.MessageRoleTool:
			tool := o.makeChatCompletionToolMessageParam(msg)
			if tool != nil {
				toolMessage := openai.ChatCompletionMessageParamUnion{OfTool: tool}
				openaiMessages = append(openaiMessages, toolMessage)
			}
		}
	}

	return openaiMessages, nil
}

func (o *openAIChat) convertToOpenAITools(tools []*llms.ToolDescriptor) ([]openai.ChatCompletionToolParam, error) {
	openaiTools := make([]openai.ChatCompletionToolParam, len(tools))
	for i, t := range tools {
		// Process function parameters
		params, err := o.convertToFunctionParameters(t)
		if err != nil {
			return nil, errors.Errorf(llms.ErrorCodeInvalidSchema,
				"failed to process parameters for function %s: %s", t.Name, err.Error())
		}

		openaiTools[i] = openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  *params,
			},
		}
	}

	return openaiTools, nil
}

func (o *openAIChat) convertToFunctionParameters(t *llms.ToolDescriptor) (*openai.FunctionParameters, error) {
	var params *openai.FunctionParameters

	if t.Parameters == nil {
		return params, nil
	}

	// Convert the schema for OpenAI compatibility
	klog.Infof("original schema for function %s: %+v", t.Name, t.Parameters)
	validatedSchema, err := o.convertSchemaForOpenAI(t.Parameters)
	if err != nil {
		return params, errors.Errorf(llms.ErrorCodeInvalidSchema,
			"schema conversion failed: %s", err.Error())
	}
	klog.Infof("converted schema for function %s: %+v", t.Name, validatedSchema)

	// Convert to raw schema bytes
	schemaBytes, err := o.convertSchemaToBytes(validatedSchema, t.Name)
	if err != nil {
		return params, err
	}

	// Unmarshal into OpenAI parameters format
	if err := json.Unmarshal(schemaBytes, &params); err != nil {
		return params, errors.Errorf(llms.ErrorCodeInvalidSchema,
			"failed to unmarshal schema: %s", err.Error())
	}

	return params, nil
}

func (o *openAIChat) makeChatCompletionParams(
	messages []*llms.Message, opts *llms.ChatOptions) (*openai.ChatCompletionNewParams, error) {
	openaiMessages, err := o.convertToOpenAIMessages(messages)
	if err != nil {
		return nil, err
	}

	params := &openai.ChatCompletionNewParams{
		Model:    o.model.ApiModelName,
		Messages: openaiMessages,
	}

	openaiTools, err := o.convertToOpenAITools(opts.Tools)
	if err != nil {
		return nil, err
	}
	if len(openaiTools) > 0 {
		params.Tools = openaiTools
	}

	if opts.Temperature != nil {
		params.Temperature = openai.Float(*opts.Temperature)
	}

	if opts.TopP != nil {
		params.TopP = openai.Float(*opts.TopP)
	}

	if opts.FrequencyPenalty != nil {
		params.FrequencyPenalty = openai.Float(*opts.FrequencyPenalty)
	}

	if opts.PresencePenalty != nil {
		params.PresencePenalty = openai.Float(*opts.PresencePenalty)
	}

	if opts.MaxCompletionTokens != nil {
		params.MaxCompletionTokens = openai.Int(*opts.MaxCompletionTokens)
	}

	if o.model.IsSupport(llms.ModelFeatureReasoning) {
		params.ReasoningEffort = o.convertToOpenAIReasoningEffort(opts.ReasoningEffort)
	} else {
		params.ReasoningEffort = o.convertToOpenAIReasoningEffort(OpenAIDefaultReasoningEffort)
	}

	return params, nil
}

func (o *openAIChat) convertToOpenAIReasoningEffort(reasoningEffort llms.ReasoningEffort) openai.ReasoningEffort {
	switch reasoningEffort {
	case llms.ReasoningEffortLow:
		return shared.ReasoningEffortLow
	case llms.ReasoningEffortMedium:
		return shared.ReasoningEffortMedium
	case llms.ReasoningEffortHigh:
		return shared.ReasoningEffortHigh
	default:
		return shared.ReasoningEffortMedium
	}
}

func (o *openAIChat) toFinishReason(reason string) llms.FinishReason {
	// Any of "stop", "length", "tool_calls", "content_filter", "function_call".
	if len(reason) == 0 {
		return ""
	}
	switch reason {
	case "stop":
		return llms.FinishReasonNormalEnd
	case "length":
		return llms.FinishReasonMaxTokens
	case "content_filter":
		return llms.FinishReasonDenied
	case "tool_calls", "function_call":
		return llms.FinishReasonToolUse
	default:
		return llms.FinishReasonUnknown
	}
}

func (o *openAIChat) getUsageStats(usage *openai.CompletionUsage) llms.UsageMetadata {
	if usage == nil {
		return llms.UsageMetadata{}
	}
	return llms.UsageMetadata{
		InputTokens:         usage.PromptTokens - usage.PromptTokensDetails.CachedTokens,
		OutputTokens:        usage.CompletionTokens,
		CacheReadTokens:     usage.PromptTokensDetails.CachedTokens,
		CacheCreationTokens: 0,
	}
}

func (o *openAIChat) toToolCalls(openaiToolCalls []openai.ChatCompletionMessageToolCall) []*llms.ToolCall {
	var toolCalls []*llms.ToolCall
	for _, tool := range openaiToolCalls {
		if len(tool.Function.Name) == 0 {
			klog.Info("skipping non-function tool call", "tool call id", tool.ID)
			continue
		}
		toolCalls = append(toolCalls, o.toToolCall(tool.ID, tool.Function))
	}
	return toolCalls
}

func (o *openAIChat) toToolCall(
	toolCallId string, toolCallFunction openai.ChatCompletionMessageToolCallFunction) *llms.ToolCall {
	var args map[string]any
	if toolCallFunction.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCallFunction.Arguments), &args); err != nil {
			klog.InfoS("error unmarshalling function arguments", "tool", toolCallFunction.Name, "err", err)
			args = make(map[string]any)
		}
	} else {
		args = make(map[string]any)
	}
	toolCall := &llms.ToolCall{
		ToolCallId: toolCallId,
		Name:       toolCallFunction.Name,
		Arguments:  args,
	}
	return toolCall
}

func (o *openAIChat) isRetryable(err error) bool {
	return errors.IsRetryableError(err)
}

func (o *openAIChat) convertSchemaForOpenAI(schema *llms.Schema) (*llms.Schema, error) {

	if schema == nil {
		// Return a minimal valid object schema for OpenAI
		return &llms.Schema{
			Type:       llms.TypeObject,
			Properties: make(map[string]*llms.Schema),
		}, nil
	}

	// Create a deep copy to avoid modifying the original
	validated := &llms.Schema{
		Description: schema.Description,
		Required:    make([]string, len(schema.Required)),
	}
	copy(validated.Required, schema.Required)

	// Handle type validation and normalization based on OpenAI requirements
	switch schema.Type {
	case llms.TypeObject:
		validated.Type = llms.TypeObject
		// Objects MUST have properties for OpenAI (even if empty)
		validated.Properties = make(map[string]*llms.Schema)
		if schema.Properties != nil {
			for key, prop := range schema.Properties {
				validatedProp, err := o.convertSchemaForOpenAI(prop)
				if err != nil {
					return nil, errors.Errorf(llms.ErrorCodeInvalidSchema,
						"validating property %q: %s", key, err.Error())
				}
				validated.Properties[key] = validatedProp
			}
		}

	case llms.TypeArray:
		validated.Type = llms.TypeArray
		// Arrays MUST have items schema for OpenAI
		if schema.Items != nil {
			validatedItems, err := o.convertSchemaForOpenAI(schema.Items)
			if err != nil {
				return nil, errors.Errorf(llms.ErrorCodeInvalidSchema,
					"validating array items: %s", err.Error())
			}
			validated.Items = validatedItems
		} else {
			// Default to string items if not specified
			validated.Items = &llms.Schema{Type: llms.TypeString}
		}

	case llms.TypeString:
		validated.Type = llms.TypeString

	case llms.TypeNumber:
		validated.Type = llms.TypeNumber

	case llms.TypeInteger:
		// OpenAI prefers "number" for integers
		validated.Type = llms.TypeNumber

	case llms.TypeBoolean:
		validated.Type = llms.TypeBoolean

	case "":
		// If no type specified, default to object with empty properties
		klog.Warningf("Schema has no type, defaulting to object")
		validated.Type = llms.TypeObject
		validated.Properties = make(map[string]*llms.Schema)

	default:
		// For unknown types, log a warning and default to object
		klog.Warningf("Unknown schema type '%s', defaulting to object", schema.Type)
		validated.Type = llms.TypeObject
		validated.Properties = make(map[string]*llms.Schema)
	}

	// Final validation: Ensure object types always have properties
	// This handles edge cases where malformed schemas might slip through
	if validated.Type == llms.TypeObject && validated.Properties == nil {
		klog.Warningf("Object schema missing properties, initializing empty properties map")
		validated.Properties = make(map[string]*llms.Schema)
	}

	return validated, nil
}

func (o *openAIChat) convertSchemaToBytes(schema *llms.Schema, toolName string) ([]byte, error) {
	// Wrap the schema with OpenAI-specific marshaling behavior
	openAIWrapper := openAISchema{Schema: schema}

	bytes, err := json.Marshal(openAIWrapper)
	if err != nil {
		return nil, errors.Errorf(llms.ErrorCodeInvalidSchema,
			"failed to convert schema: %s", err.Error())
	}

	klog.Infof("- OpenAI schema for function %s: %s", toolName, string(bytes))

	return bytes, nil
}

func (o *openAIChat) makeChatCompletionContent(msg *llms.Message) []openai.ChatCompletionContentPartUnionParam {
	var content []openai.ChatCompletionContentPartUnionParam
	for _, part := range msg.Parts {
		switch part.Type() {
		case llms.PartTypeText:
			if realPart, ok := part.(*llms.TextPart); ok {
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: realPart.Text,
					},
				})
			}
		case llms.PartTypeData:
			if realPart, ok := part.(*llms.DataPart); ok {
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: fmt.Sprintf("\n```\n%s\n```\n", realPart.MarshalJson()),
					},
				})
			}
		case llms.PartTypeBinary:
			if realPart, ok := part.(*llms.BinaryPart); ok {
				if format, is := llms.IsImagePart(realPart); is {
					if len(realPart.Content) > 0 {
						base64Content := realPart.MarshalBase64()
						content = append(content, openai.ChatCompletionContentPartUnionParam{
							OfImageURL: &openai.ChatCompletionContentPartImageParam{
								ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
									URL: fmt.Sprintf("data:%s;base64,%s", realPart.MIMEType, base64Content),
								},
							},
						})
					} else if realPart.URL != nil {
						content = append(content, openai.ChatCompletionContentPartUnionParam{
							OfImageURL: &openai.ChatCompletionContentPartImageParam{
								ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
									URL: *realPart.URL,
								},
							},
						})
					} else {
						// No-ops
					}
				} else if format, is = llms.IsAudioPart(realPart); is {
					content = append(content, openai.ChatCompletionContentPartUnionParam{
						OfInputAudio: &openai.ChatCompletionContentPartInputAudioParam{
							InputAudio: openai.ChatCompletionContentPartInputAudioInputAudioParam{
								Data:   realPart.MarshalBase64(),
								Format: format,
							},
						},
					})
				} else {
					fileName := ""
					if realPart.Name != nil {
						fileName = *realPart.Name
					}
					fileId := ""
					if realPart.URL != nil {
						fileId = *realPart.URL
					}
					if len(realPart.Content) > 0 {
						base64Content := realPart.MarshalBase64()
						content = append(content, openai.ChatCompletionContentPartUnionParam{
							OfFile: &openai.ChatCompletionContentPartFileParam{
								File: openai.ChatCompletionContentPartFileFileParam{
									Filename: openai.String(fileName),
									FileID:   openai.String(fileId),
									FileData: openai.String(base64Content),
								},
							},
						})
					} else {
						lines := []string{"> reference file:"}
						if len(fileName) > 0 {
							lines = append(lines, fmt.Sprintf("> - file name: %s", fileName))
						}
						if len(fileId) > 0 {
							lines = append(lines, fmt.Sprintf("> - url: %s", fileId))
						}
						if len(lines) > 1 {
							content = append(content, openai.ChatCompletionContentPartUnionParam{
								OfText: &openai.ChatCompletionContentPartTextParam{
									Text: strings.Join(lines, "\n"),
								},
							})
						}
					}
				}
			}
		}
	}
	return content
}

func (o *openAIChat) makeChatCompletionAssistantMessageParam(msg *llms.Message) *openai.ChatCompletionAssistantMessageParam {
	var content []openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion
	var toolCalls []openai.ChatCompletionMessageToolCallParam
	for _, part := range msg.Parts {
		switch part.Type() {
		case llms.PartTypeText:
			if realPart, ok := part.(*llms.TextPart); ok {
				content = append(content, openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: realPart.Text,
					},
				})
			}
		case llms.PartTypeToolCall:
			if realPart, ok := part.(*llms.ToolCall); ok {
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
					ID: realPart.ToolCallId,
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      realPart.Name,
						Arguments: realPart.MarshalJson(),
					},
				})
			}
		}
	}
	return &openai.ChatCompletionAssistantMessageParam{
		Content: openai.ChatCompletionAssistantMessageParamContentUnion{
			OfArrayOfContentParts: content,
		},
		ToolCalls: toolCalls,
	}
}

func (o *openAIChat) makeChatCompletionToolMessageParam(msg *llms.Message) *openai.ChatCompletionToolMessageParam {
	var tool openai.ChatCompletionToolMessageParam

	for _, part := range msg.Parts {
		if part.Type() != llms.PartTypeToolCallResult {
			continue
		}
		realPart, ok := part.(*llms.ToolCallResult)
		if !ok {
			continue
		}

		tool.ToolCallID = realPart.ToolCallId
		tool.Content = openai.ChatCompletionToolMessageParamContentUnion{
			OfString: openai.String(realPart.MarshalJson()),
		}

		break
	}

	return &tool
}

func (o *openAIChat) makeSystemMessage(messages []*llms.Message) openai.ChatCompletionMessageParamUnion {
	msg := llms.MakeSystemInstruction(o.systemPrompt, messages)
	return openai.SystemMessage(msg)
}

// openAISchema wraps a Schema with OpenAI-specific marshaling behavior
type openAISchema struct {
	*llms.Schema
}

// MarshalJSON provides OpenAI-specific JSON marshaling that ensures object schemas have properties
func (s openAISchema) MarshalJSON() ([]byte, error) {
	// Create a map to build the JSON representation
	result := make(map[string]interface{})

	if s.Type != "" {
		result["type"] = s.Type
	}

	if s.Description != "" {
		result["description"] = s.Description
	}

	if len(s.Required) > 0 {
		result["required"] = s.Required
	}

	// For object types, always include properties (even if empty) to satisfy OpenAI
	if s.Type == llms.TypeObject {
		if s.Properties != nil {
			result["properties"] = s.Properties
		} else {
			result["properties"] = make(map[string]*llms.Schema)
		}
	} else if s.Properties != nil && len(s.Properties) > 0 {
		// For non-object types, only include properties if they exist and are non-empty
		result["properties"] = s.Properties
	}

	if s.Items != nil {
		result["items"] = s.Items
	}

	return json.Marshal(result)
}
