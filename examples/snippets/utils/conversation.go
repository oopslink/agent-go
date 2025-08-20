package utils

import (
	"context"
	"fmt"
	"github.com/oopslink/agent-go/pkg/commons/errors"
	"github.com/oopslink/agent-go/pkg/commons/utils"
	"github.com/oopslink/agent-go/pkg/core/agent"
)

// RunSingleConversation runs a single conversation with the agent
func RunSingleConversation(theAgent agent.Agent, loading, question string) error {
	ctx := context.Background()

	// Start the agent
	inputChan, outputChan, err := theAgent.Run(
		&agent.RunContext{
			SessionId: fmt.Sprintf("cot-%s", utils.GenerateUUID()),
			Context:   ctx,
		})
	if err != nil {
		return errors.Errorf(errors.InternalError, "failed to start agent: %v", err)
	}

	// Send the question
	inputChan <- agent.NewUserRequestEvent(
		&agent.UserRequest{
			Message: question,
		})

	// Process response
	fmt.Printf("Question: %s\n\n", question)
	fmt.Println(loading)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-outputChan:
			if !ok {
				return nil
			}
			done, err := HandleAgentEvent(event)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}
