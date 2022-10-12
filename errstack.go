package funcchain

import (
	"sync"
)

// ErrorStack a stack of error
type ErrorStack struct {
	errs []error
	lock sync.RWMutex
}

// NewErrorStack create a new ErrorStack
func NewErrorStack() *ErrorStack {
	return &ErrorStack{
		errs: []error{},
	}
}

func (s *ErrorStack) Push(err error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.errs = append(s.errs, err)
}

func (s *ErrorStack) Pop() (error, bool) {
	if s.IsEmpty() {
		return nil, false
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	err := s.errs[len(s.errs)-1]
	s.errs = s.errs[0 : len(s.errs)-1]
	return err, true
}

func (s *ErrorStack) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.errs = []error{}

}

func (s *ErrorStack) GetAll() []error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.errs
}

func (s *ErrorStack) Size() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.errs)
}

func (s *ErrorStack) IsEmpty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.errs) == 0
}
