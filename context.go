package funcchain

import (
	"context"
	"math/bits"
	"sync"
	"time"
)

const (
	Abort     FuncState = (1 << bits.UintSize) / -2
	Undefined FuncState = (1<<bits.UintSize)/-2 + 1
)

type FuncState int

type Context struct {
	context.Context
	mu    sync.RWMutex
	meta  sync.Map
	state FuncState
}

func Init(ctx context.Context) *Context {
	return &Context{
		Context: ctx,
		state:   Undefined,
	}
}

func (c *Context) SetState(state FuncState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = state
}

func (c *Context) GetState() FuncState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Context) Set(key string, value interface{}) {
	c.meta.Store(key, value)
}

func (c *Context) Get(key string) (value interface{}, ok bool) {
	return c.meta.Load(key)
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.Context.Done()
}

func (c *Context) Err() error {
	return c.Context.Err()
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}
