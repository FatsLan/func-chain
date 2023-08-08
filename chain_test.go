package funcchain

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

// TODO 1:函数链实现重试 retry
// TODO 2:实现注入函数判断是否需要执行

func metricsHandler(chain *HandlerChain, handler Handler) error {
	//t := time.Now().UnixNano()
	chain.Do(handler)
	//fmt.Printf("test metrics:%v, cost:%v\n", handler.options.GetSign(), time.Now().UnixNano()-t)

	return nil
}

func controllerHandler(ctx *Context, state FuncState) ControlFlag {
	if ctx.state == StateAbort {
		return Break
	}
	if state > ctx.state {
		return Continue
	}
	return GoOn
}

func func1(ctx *Context) error {
	//fmt.Printf("test 1\n")
	return errors.New("1")
}

func func2(ctx *Context) error {
	//fmt.Printf("test 2\n")
	return errors.New("2")
}

func func3(ctx *Context) error {
	v, _ := ctx.Get("k")
	if v.(int) < 3 {
		ctx.Set("k", v.(int)+1)
		fmt.Printf("test 3 %d\n", v.(int))
		return RetryThenAbort(ctx, errors.New("3"))
	}
	fmt.Printf("stop retry, test 3 %d\n", v.(int))
	return errors.New("3")
}

func func4(ctx *Context) error {

	return errors.New("4")
}

func funcAbort(ctx *Context) error {
	ctx.SetState(StateAbort)
	fmt.Printf("test abort\n")
	return nil
}

// 从中间状态开始执行
func Test(t *testing.T) {

	ctx := context.Background()
	chain := NewHandlerChain(ctx).WithState(1)
	//chain.RegisterControllerHandler(controllerHandler)
	chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(func3, WithSign("func3"), WithState(3))
	chain.Use(func4, WithSign("func4"), WithState(4), WithRetryCount(10))

	err := chain.Run()
	if err != nil {
		fmt.Printf("err:%v\n", err)
		e, _ := chain.errs.Pop()
		fmt.Printf("err:%v\n", e)
		e, _ = chain.errs.Pop()
		fmt.Printf("err:%v\n", e)
		e, _ = chain.errs.Pop()
		fmt.Printf("err:%v\n", e)
		//fmt.Printf("err:%v\n", chain.ErrorAll())
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
	chain := NewHandlerChain(ctx)
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

// 测试 retry
func Test5(t *testing.T) {
	ctx := context.Background()
	chain := NewHandlerChain(ctx)
	chain.Set("k", 1)
	//chain.RegisterControllerHandler(controllerHandler)
	chain.RegisterMetricsHandler(metricsHandler)

	chain.Use(func1, WithSign("func1"), WithState(1))
	chain.Use(func2, WithSign("func2"), WithState(2))
	chain.Use(func3, WithSign("func3"), WithState(3), WithRetryCount(5))
	chain.Use(func4, WithSign("func4"), WithState(4))

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
		err := func1(ctx)
		err = func2(ctx)
		err = func3(ctx)
		err = func4(ctx)
		if err != nil {
			fmt.Printf("err:%v\n", err)
		}
	}
}
