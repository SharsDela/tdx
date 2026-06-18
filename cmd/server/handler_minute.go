package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/injoyai/tdx"
)

func (s *Server) handleMinute(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("20060102")
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetHistoryMinute(date, code) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleTrade(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	start := uint16(0)
	count := uint16(100)
	if s := r.URL.Query().Get("start"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 16); err == nil {
			start = uint16(v)
		}
	}
	if c := r.URL.Query().Get("count"); c != "" {
		if v, err := strconv.ParseUint(c, 10, 16); err == nil {
			count = uint16(v)
		}
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetMinuteTrade(code, start, count) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleHistoryTrade(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	date := r.URL.Query().Get("date")
	if date == "" {
		writeError(w, http.StatusBadRequest, "date parameter required (format: 20060102)")
		return
	}
	start := uint16(0)
	count := uint16(100)
	if s := r.URL.Query().Get("start"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 16); err == nil {
			start = uint16(v)
		}
	}
	if c := r.URL.Query().Get("count"); c != "" {
		if v, err := strconv.ParseUint(c, 10, 16); err == nil {
			count = uint16(v)
		}
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) {
		return c.GetHistoryMinuteTrade(date, code, start, count)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
