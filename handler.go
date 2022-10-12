package funcchain

const (
	GoOn ControlFlag = iota
	Continue
	Break
)

type ControlFlag int

type MetricsHandler func(chain *HandlerChain, handler Handler) error

type ErrorHandler func(ctx *Context, err error) error

type FuncHandler func(ctx *Context) (ErrorHandler, error)

type ControllerHandler func(ctx *Context, state FuncState) ControlFlag

// HandlerOption 可选参数
type HandlerOption struct {
	state FuncState // func 状态
	sign  string    // func 标记
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

func (h *HandlerOption) GetState() FuncState {
	return h.state
}

func (h *HandlerOption) GetSign() string {
	return h.sign
}

type Handler struct {
	funcHandler FuncHandler
	options     *HandlerOption
}
