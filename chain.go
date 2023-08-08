package funcchain

import (
	"context"
	"math"

	"github.com/pkg/errors"
)

const maxLen = math.MaxInt8

type HandlerChain struct {
	*Context
	errs              *ErrorStack
	handlers          []Handler
	metricsHandler    MetricsHandler
	controllerHandler ControllerHandler
}

func NewHandlerChain(ctx context.Context) *HandlerChain {
	return &HandlerChain{
		Context: Init(ctx),
		errs:    NewErrorStack(),
	}
}

func (handlerChain *HandlerChain) WithState(state FuncState) *HandlerChain {
	handlerChain.Context.SetState(state)
	return handlerChain
}

func (handlerChain *HandlerChain) Use(f FuncHandler, opts ...Options) {
	handler := Handler{
		funcHandler: f,
		options:     &HandlerOption{},
	}

	for _, op := range opts {
		op(handler.options)
	}

	handlerChain.handlers = append(handlerChain.handlers, handler)
}

// Abort handle error info
func Abort(ctx *Context, err error) error {
	ctx.Abort()
	return err
}

// RetryThenAbort retry Handler, if can't get except response after retry count reduce to 0, handle error info
func RetryThenAbort(ctx *Context, err error) error {
	ctx.RetryThenAbort()
	return err
}

// RegisterMetricsHandler register the common MetricsHandler in HandlerChain
func (handlerChain *HandlerChain) RegisterMetricsHandler(handler MetricsHandler) {
	handlerChain.metricsHandler = handler
}

func (handlerChain *HandlerChain) RegisterControllerHandler(handler ControllerHandler) {
	handlerChain.controllerHandler = handler
}

// Run func in func chain, when the function is done,
// HandlerChain will handle the callback by the rule.
// When func’s callback is registered, execute it preferential.
// When func’s callback is nil, it will execute HandlerChain’s common errorHandler.
// When HandlerChain’s common errorHandler is also nil, it will pop the original error.
func (handlerChain *HandlerChain) Run() error {
	if len(handlerChain.handlers) > maxLen {
		return errors.WithStack(errors.New("out of max length"))
	}

Loop:
	for _, h := range handlerChain.handlers {
		retry := h.options.GetRetryCount()

		handlerChain.do(h)

		flag := doDefaultController(handlerChain, &h.options.state)
		// 如果 func 需要 retry，那么这里需要重定向到上一个 handlers
		for flag == Retry && retry > 0 {
			handlerChain.ClearState()
			handlerChain.do(h)
			flag = doDefaultController(handlerChain, &h.options.state)
			retry--
		}

		switch flag {
		case Retry, Break:
			break Loop
		case Continue:
			continue
		default:
		}

	}

	err, _ := handlerChain.PopError()
	return err
}

func (handlerChain *HandlerChain) Do(h Handler) {
	handlerChain.doDefault(h)
}

func (handlerChain *HandlerChain) doDefault(h Handler) {
	handlerChain.handleError(h.funcHandler(handlerChain.Context))
}

func (handlerChain *HandlerChain) handleError(err error) {
	if err == nil {
		return
	}

	handlerChain.errs.Push(err)
}

func (handlerChain *HandlerChain) pushError(err error) {
	handlerChain.errs.Push(err)
}

func (handlerChain *HandlerChain) PopError() (error, bool) {
	return handlerChain.errs.Pop()
}

func (handlerChain *HandlerChain) ErrorAll() []error {
	return handlerChain.errs.GetAll()
}

func doDefaultController(handlerChain *HandlerChain, state *FuncState) ControlFlag {
	if handlerChain.Context.GetState() == StateAbort {
		return Break
	}

	if handlerChain.Context.GetState() == StateRetryThenAbort {
		return Retry
	}

	if state != nil {
		if handlerChain.Context.GetState() > *state {
			return Continue
		}

		handlerChain.Context.SetState(*state)
	}
	return GoOn
}

func (handlerChain *HandlerChain) do(h Handler) {
	if handlerChain.metricsHandler != nil {
		_ = handlerChain.metricsHandler(handlerChain, h)
	} else {
		handlerChain.doDefault(h)
	}
}
