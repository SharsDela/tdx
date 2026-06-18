package main

import (
	"net/http"
	"strconv"

	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

var klineTypes = map[string]uint8{
	"minute":  protocol.TypeKlineMinute,
	"5min":    protocol.TypeKline5Minute,
	"15min":   protocol.TypeKline15Minute,
	"30min":   protocol.TypeKline30Minute,
	"60min":   protocol.TypeKline60Minute,
	"day":     protocol.TypeKlineDay,
	"week":    protocol.TypeKlineWeek,
	"month":   protocol.TypeKlineMonth,
	"quarter": protocol.TypeKlineQuarter,
	"year":    protocol.TypeKlineYear,
}

func parseKlineParams(r *http.Request) (ktype uint8, start, count uint16) {
	ktype = protocol.TypeKlineDay
	if t := r.URL.Query().Get("type"); t != "" {
		if v, ok := klineTypes[t]; ok {
			ktype = v
		}
	}
	start = 0
	if s := r.URL.Query().Get("start"); s != "" {
		if v, err := strconv.ParseUint(s, 10, 16); err == nil {
			start = uint16(v)
		}
	}
	count = 100
	if c := r.URL.Query().Get("count"); c != "" {
		if v, err := strconv.ParseUint(c, 10, 16); err == nil {
			count = uint16(v)
		}
	}
	return
}

func (s *Server) handleKline(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	ktype, start, count := parseKlineParams(r)
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetKline(ktype, code, start, count) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleIndexKline(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	ktype, start, count := parseKlineParams(r)
	result, err := s.clientDo(func(c *tdx.Client) (any, error) { return c.GetIndex(ktype, code, start, count) })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
