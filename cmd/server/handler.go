package main

import (
	"time"

	"github.com/injoyai/tdx"
)

type Server struct {
	pool    tdx.IPool
	timeout time.Duration
}

func (s *Server) clientDo(fn func(c *tdx.Client) (any, error)) (any, error) {
	c, err := s.pool.Get()
	if err != nil {
		return nil, err
	}
	defer s.pool.Put(c)
	c.SetTimeout(s.timeout)
	return fn(c)
}
