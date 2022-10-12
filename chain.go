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
	errorHandler      ErrorHandler
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

// RegisterErrorHandler register the common ErrorHandler in HandlerChain
func (handlerChain *HandlerChain) RegisterErrorHandler(handler ErrorHandler) {
	handlerChain.errorHandler = handler
}

func (handlerChain *HandlerChain) RegisterMetricsHandler(handler MetricsHandler) {
	handlerChain.metricsHandler = handler
}

// RegisterBizHandler RegisterBizHandler
func (handlerChain *HandlerChain) RegisterBizHandler(handler ErrorHandler) {
	handlerChain.errorHandler = handler
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
		flag := doDefaultController(handlerChain, &h.options.state)

		switch flag {
		case Break:
			break Loop
		case Continue:
			continue
		default:
		}

		if handlerChain.metricsHandler != nil {
			_ = handlerChain.metricsHandler(handlerChain, h)
		} else {
			handlerChain.doDefault(h)
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

func (handlerChain *HandlerChain) handleError(handler ErrorHandler, err error) {
	if err == nil {
		return
	}

	// 先执行函数绑定的回调函数
	if handler != nil {
		handlerChain.errs.Push(handler(handlerChain.Context, err))
		return
	}

	// 再执行 chain 注册的公共回调函数
	if handlerChain.errorHandler != nil {
		handlerChain.errs.Push(handlerChain.errorHandler(handlerChain.Context, err))
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
	if handlerChain.Context.GetState() == Abort {
		return Break
	}

	if state != nil {
		if handlerChain.Context.GetState() > *state {
			return Continue
		}

		handlerChain.Context.SetState(*state)
	}
	return GoOn
}
