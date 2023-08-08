package funcchain

const (
	GoOn     ControlFlag = iota
	Continue             // Skip the next function when state = Continue
	Retry                // Retry the Handler when state = Retry
	Break                // Abort and exit from the function chain when state = Abort
)

type ControlFlag int

type MetricsHandler func(chain *HandlerChain, handler Handler) error

type ErrorHandler func(ctx *Context, err error) error

type FuncHandler func(ctx *Context) error

type ControllerHandler func(ctx *Context, state FuncState) ControlFlag

// HandlerOption 可选参数
type HandlerOption struct {
	state      FuncState // func 状态
	sign       string    // func 标记
	retryCount int       // retry count
}

type Options func(*HandlerOption)

func WithState(state FuncState) Options {
	return func(o *HandlerOption) {
		o.state = state
	}
}

func WithSign(sign string) Options {
	return func(o *HandlerOption) {
		o.sign = sign
	}
}

// WithRetryCount set retry count, then Handler
// will retry when FuncState is Abort and retry count > 0
func WithRetryCount(retry int) Options {
	return func(o *HandlerOption) {
		o.retryCount = retry
	}
}

func (h *HandlerOption) GetState() FuncState {
	return h.state
}

func (h *HandlerOption) GetSign() string {
	return h.sign
}

func (h *HandlerOption) GetRetryCount() int {
	return h.retryCount
}

type Handler struct {
	funcHandler FuncHandler
	options     *HandlerOption
}
