package eventbus

import "github.com/oopslink/agent-go/pkg/commons/errors"

var (
	ErrorCodeEventBusAlreadyClosed = errors.ErrorCode{
		Code:           30500,
		Name:           "EventBusAlreadyClosed ",
		DefaultMessage: "Event bus already closed",
	}
	ErrorCodeSubscriberAlreadyClosed = errors.ErrorCode{
		Code:           30501,
		Name:           "SubscriberAlreadyClosed ",
		DefaultMessage: "Event bus subscriber already closed",
	}
)
