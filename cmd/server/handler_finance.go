package main

import (
	"net/http"
	"strconv"

	"github.com/injoyai/tdx"
)

func (s *Server) handleFinance(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	ex := r.URL.Query().Get("exchange")
	if ex == "" {
		ex = defaultExchange(code)
	}
	exchange, ok := parseExchange(ex)
	if !ok {
		writeError(w, http.StatusBadRequest, "exchange parameter required: sh, sz, bj")
		return
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetFinanceInfo(exchange, code) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCompanyCategory(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	ex := r.URL.Query().Get("exchange")
	if ex == "" {
		ex = defaultExchange(code)
	}
	exchange, ok := parseExchange(ex)
	if !ok {
		writeError(w, http.StatusBadRequest, "exchange parameter required: sh, sz, bj")
		return
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetCompanyCategory(exchange, code) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCompanyContent(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	ex := r.URL.Query().Get("exchange")
	if ex == "" {
		ex = defaultExchange(code)
	}
	exchange, ok := parseExchange(ex)
	if !ok {
		writeError(w, http.StatusBadRequest, "exchange parameter required: sh, sz, bj")
		return
	}
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeError(w, http.StatusBadRequest, "filename parameter required")
		return
	}
	start := uint32(0)
	length := uint32(0xFFFF)
	if s := r.URL.Query().Get("start"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 32); err == nil {
			start = uint32(v)
		}
	}
	if l := r.URL.Query().Get("length"); l != "" {
		if v, err := strconv.ParseUint(l, 10, 32); err == nil {
			length = uint32(v)
		}
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) {
		return c.GetCompanyContent(exchange, code, filename, start, length)
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"content": result})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func defaultExchange(code string) string {
	if len(code) >= 2 {
		switch code[:2] {
		case "60", "68", "90", "51":
			return "sh"
		case "00", "30", "15", "16":
			return "sz"
		case "92", "83", "87", "43":
			return "bj"
		}
	}
	return "sh"
}
