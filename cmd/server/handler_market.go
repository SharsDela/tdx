package main

import (
	"net/http"
	"strings"

	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

func (s *Server) handleQuote(w http.ResponseWriter, r *http.Request) {
	codesStr := r.URL.Query().Get("codes")
	if codesStr == "" {
		writeError(w, http.StatusBadRequest, "codes parameter required")
		return
	}
	codes := strings.Split(codesStr, ",")
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetQuote(codes...) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCallAuction(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetCallAuction(code) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGbbq(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetGbbq(code) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if file == "" {
		writeError(w, http.StatusBadRequest, "file parameter required")
		return
	}
	if !strings.HasPrefix(file, "block_") {
		writeError(w, http.StatusBadRequest, "file must start with block_")
		return
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetBlockData(file) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleBlockWithIndex(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if file == "" {
		writeError(w, http.StatusBadRequest, "file parameter required")
		return
	}
	if !strings.HasPrefix(file, "block_") {
		writeError(w, http.StatusBadRequest, "file must start with block_")
		return
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetBlockDataWithIndex(file) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleXgsg(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetXgsg() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleStat(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetTdxStat() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleStat2(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetTdxStat2() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleCodes(w http.ResponseWriter, r *http.Request) {
	ex := r.PathValue("exchange")
	exchange, ok := parseExchange(ex)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid exchange: use sh, sz, bj")
		return
	}
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetCodeAll(exchange) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAllStocks(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetStockCodeAll() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"codes": result})
}

func (s *Server) handleAllETFs(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetETFCodeAll() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"codes": result})
}

func (s *Server) handleAllIndexes(w http.ResponseWriter, r *http.Request) {
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetIndexCodeAll() })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"codes": result})
}

func parseExchange(s string) (protocol.Exchange, bool) {
	switch s {
	case "sh":
		return protocol.ExchangeSH, true
	case "sz":
		return protocol.ExchangeSZ, true
	case "bj":
		return protocol.ExchangeBJ, true
	default:
		return 0, false
	}
}
