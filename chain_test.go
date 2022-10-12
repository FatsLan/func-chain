package funcchain

import (
	"context"
	"fmt"
	"testing"
)

func metricsHandler(chain *HandlerChain, handler Handler) error {
	//t := time.Now().UnixNano()
	chain.Do(handler)
	//fmt.Printf("test metrics:%v, cost:%v\n", handler.options.GetSign(), time.Now().UnixNano()-t)

	return nil
}

func controllerHandler(ctx *Context, state FuncState) ControlFlag {
	if ctx.state == Abort {
		return Break
	}
	if state > ctx.state {
		return Continue
	}
	return GoOn
}

func func1(ctx *Context) (ErrorHandler, error) {
	//fmt.Printf("test 1\n")
	return nil, nil
}

func func2(ctx *Context) (ErrorHandler, error) {
	//fmt.Printf("test 2\n")
	return nil, nil
}

func func3(ctx *Context) (ErrorHandler, error) {
	//fmt.Printf("test 3\n")
	return nil, nil
}

func func4(ctx *Context) (ErrorHandler, error) {
	//fmt.Printf("test 4\n")
	return nil, nil
}

func funcAbort(ctx *Context) (ErrorHandler, error) {
	ctx.SetState(Abort)
	fmt.Printf("test abort\n")
	return nil, nil
}

// 从中间状态开始执行
func Test(t *testing.T) {

	ctx := context.Background()
	chain := NewHandlerChain(ctx).WithState(3)
	//chain.RegisterControllerHandler(controllerHandler)
	chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(func3, WithSign("func3"), WithState(3))
	chain.Use(func4, WithSign("func4"), WithState(4))

	err := chain.Run()
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

// 从开头开始执行
func Test2(t *testing.T) {
	ctx := context.Background()
	chain := NewHandlerChain(ctx).WithState(1)
	//chain.RegisterControllerHandler(controllerHandler)
	//chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(func3, WithSign("func3"), WithState(3))
	chain.Use(func4, WithSign("func4"), WithState(4))

	err := chain.Run()
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

// 执行一半退出
func Test3(t *testing.T) {
	ctx := context.Background()
	chain := NewHandlerChain(ctx).WithState(1)
	//chain.RegisterControllerHandler(controllerHandler)
	//chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(funcAbort, WithSign("funcAbort"), WithState(3))
	chain.Use(func3, WithSign("func3"), WithState(4))
	chain.Use(func4, WithSign("func4"), WithState(5))

	err := chain.Run()
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

// 测试走 metrics
func Test4(t *testing.T) {
	ctx := context.Background()
	chain := NewHandlerChain(ctx).WithState(1)
	//chain.RegisterControllerHandler(controllerHandler)
	chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(func3, WithSign("func3"), WithState(4))
	chain.Use(func4, WithSign("func4"), WithState(5))

	err := chain.Run()
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
}

func Benchmark1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := context.Background()
		chain := NewHandlerChain(ctx).WithState(1)
		//chain.RegisterControllerHandler(controllerHandler)
		//chain.RegisterMetricsHandler(metricsHandler)

		chain.Use(func1)
		chain.Use(func2)
		chain.Use(func3)
		chain.Use(func4)

		err := chain.Run()
		if err != nil {
			fmt.Printf("err:%v\n", err)
		}
	}
}

func Benchmark2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ctx := Init(context.Background())
		errHandler, err := func1(ctx)
		if err != nil {
			_ = errHandler(ctx, err)
		}

		errHandler, err = func2(ctx)
		if err != nil {
			_ = errHandler(ctx, err)
		}

		errHandler, err = func3(ctx)
		if err != nil {
			_ = errHandler(ctx, err)
		}

		errHandler, err = func4(ctx)
		if err != nil {
			_ = errHandler(ctx, err)
		}
	}
}
