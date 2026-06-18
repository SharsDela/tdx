package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/injoyai/tdx"
)

func main() {
	cfg := loadConfig()

	pool, err := tdx.NewPool(func() (*tdx.Client, error) {
		return tdx.DialHostsRange(cfg.Hosts, tdx.WithDebug())
	}, cfg.PoolSize)
	if err != nil {
		log.Fatal("failed to create connection pool:", err)
	}

	log.Printf("tdx hosts: %v", cfg.Hosts)

	// 初始化 DefaultCodes（GetQuote 处理非股票代码 ETF/指数 时必需）
	// 使用独立连接，避免占用 pool 名额
	codesClient, err := tdx.DialHostsRange(cfg.Hosts)
	if err != nil {
		log.Printf("[WARN] failed to dial codes client: %v (ETF/index quote will fail)", err)
	} else {
		codes, err := tdx.NewCodesSqlite(tdx.WithCodesClient(codesClient))
		if err != nil {
			log.Printf("[WARN] failed to init DefaultCodes: %v (ETF/index quote will fail)", err)
		} else {
			tdx.DefaultCodes = codes
			log.Printf("DefaultCodes initialized OK")
		}
	}

	srv := &Server{pool: pool, timeout: cfg.Timeout}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/quote", srv.handleQuote)
	mux.HandleFunc("GET /api/v1/kline/{code}", srv.handleKline)
	mux.HandleFunc("GET /api/v1/index/kline/{code}", srv.handleIndexKline)
	mux.HandleFunc("GET /api/v1/minute/{code}", srv.handleMinute)
	mux.HandleFunc("GET /api/v1/trade/{code}", srv.handleTrade)
	mux.HandleFunc("GET /api/v1/history/trade/{code}", srv.handleHistoryTrade)
	mux.HandleFunc("GET /api/v1/finance/{code}", srv.handleFinance)
	mux.HandleFunc("GET /api/v1/company/category/{code}", srv.handleCompanyCategory)
	mux.HandleFunc("GET /api/v1/company/content/{code}", srv.handleCompanyContent)
	mux.HandleFunc("GET /api/v1/codes/{exchange}", srv.handleCodes)
	mux.HandleFunc("GET /api/v1/codes/all/stocks", srv.handleAllStocks)
	mux.HandleFunc("GET /api/v1/codes/all/etfs", srv.handleAllETFs)
	mux.HandleFunc("GET /api/v1/codes/all/indexes", srv.handleAllIndexes)
	mux.HandleFunc("GET /api/v1/block/{file}", srv.handleBlock)
	mux.HandleFunc("GET /api/v1/block-with-index/{file}", srv.handleBlockWithIndex)
	mux.HandleFunc("GET /api/v1/gbbq/{code}", srv.handleGbbq)
	mux.HandleFunc("GET /api/v1/xgsg", srv.handleXgsg)
	mux.HandleFunc("GET /api/v1/stat", srv.handleStat)
	mux.HandleFunc("GET /api/v1/stat2", srv.handleStat2)
	mux.HandleFunc("GET /api/v1/call-auction/{code}", srv.handleCallAuction)
	mux.HandleFunc("GET /api/v1/health", srv.handleHealth)

	handler := middlewareLogging(middlewareCORS(middlewareRecovery(mux)))

	httpServer := &http.Server{
		Addr:         cfg.Listen,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("tdx-api listening on %s", cfg.Listen)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
}

type Config struct {
	Listen   string
	PoolSize int
	Timeout  time.Duration
	Hosts    []string
	Port     string
}

func loadConfig() Config {
	cfg := Config{
		Listen:   ":8080",
		PoolSize: 3,
		Timeout:  8 * time.Second,
		Port:     "7709",
	}
	if v := os.Getenv("LISTEN"); v != "" {
		cfg.Listen = v
	}
	if v := os.Getenv("POOL_SIZE"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			cfg.PoolSize = n
		}
	}
	if v := os.Getenv("TDX_PORT"); v != "" {
		cfg.Port = v
	}

	// 主机选择优先级: HOSTS 环境变量 > 交互选择(TTY) > 全部默认主机
	switch {
	case os.Getenv("HOSTS") != "":
		cfg.Hosts = strings.Split(os.Getenv("HOSTS"), ",")
	case isTTY() && os.Getenv("NO_INTERACTIVE") == "":
		cfg.Hosts = selectHosts(cfg.Port)
	default:
		cfg.Hosts = tdx.Hosts
	}

	for i, host := range cfg.Hosts {
		if !strings.Contains(host, ":") {
			cfg.Hosts[i] = host + ":" + cfg.Port
		}
	}
	return cfg
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
