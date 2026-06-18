package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/tdx"
)

type hostGroup struct {
	name  string
	hosts []string
}

type hostItem struct {
	idx   int
	group string
	host  string
}

func allHostGroups() []hostGroup {
	return []hostGroup{
		{"上海(SH)", tdx.SHHosts},
		{"北京(BJ)", tdx.BJHosts},
		{"广州(GZ)", tdx.GZHosts},
		{"武汉(WH)", tdx.WHHosts},
	}
}

// isTTY 判断 stdin 是否是终端
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// selectHosts 启动时交互式选择服务器
// 返回选中的 host 列表(不含端口)
func selectHosts(port string) []string {
	groups := allHostGroups()

	// 构建编号到host的扁平映射
	var items []hostItem
	idx := 1
	for _, g := range groups {
		for _, h := range g.hosts {
			items = append(items, hostItem{idx, g.name, h})
			idx++
		}
	}

	for {
		fmt.Println()
		fmt.Println("==================== 通达信服务器列表 ====================")
		currentGroup := ""
		for _, it := range items {
			if it.group != currentGroup {
				fmt.Printf("\n  [%s]\n", it.group)
				currentGroup = it.group
			}
			fmt.Printf("    %2d) %s\n", it.idx, it.host)
		}
		fmt.Println()
		fmt.Println("操作选项:")
		fmt.Println("  a       使用全部服务器(按顺序尝试,推荐)")
		fmt.Println("  s       测速并选择最快的 3 个")
		fmt.Println("  数字    指定编号,逗号分隔(例: 1,3,5)")
		fmt.Print("\n请选择 [a/s/数字] (默认 a): ")

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("读取输入失败,使用全部服务器")
			return tdx.Hosts
		}
		input := strings.TrimSpace(line)

		switch {
		case input == "" || input == "a" || input == "A":
			return tdx.Hosts

		case input == "s" || input == "S":
			selected := speedTestAndPick(tdx.Hosts, port, 3)
			if len(selected) == 0 {
				fmt.Println("测速未找到可用服务器,请重新选择")
				continue
			}
			return selected

		default:
			selected := parseNumberSelection(input, items)
			if len(selected) == 0 {
				fmt.Println("输入无效,请重新选择")
				continue
			}
			return selected
		}
	}
}

func parseNumberSelection(input string, items []hostItem) []string {
	parts := strings.Split(input, ",")
	var hosts []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > len(items) {
			return nil
		}
		hosts = append(hosts, items[n-1].host)
	}
	return hosts
}

// speedTestAndPick 测速所有服务器,返回最快的 N 个
func speedTestAndPick(hosts []string, port string, n int) []string {
	fmt.Printf("\n正在测试 %d 个服务器的连接速度...\n", len(hosts))

	type result struct {
		host    string
		latency time.Duration
		ok      bool
	}

	results := make([]result, len(hosts))
	var wg sync.WaitGroup
	for i, h := range hosts {
		wg.Add(1)
		go func(i int, h string) {
			defer wg.Done()
			addr := h
			if !strings.Contains(addr, ":") {
				addr += ":" + port
			}
			start := time.Now()
			c, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err != nil {
				results[i] = result{host: h, ok: false}
				return
			}
			c.Close()
			results[i] = result{host: h, latency: time.Since(start), ok: true}
		}(i, h)
	}
	wg.Wait()

	var ok []result
	for _, r := range results {
		if r.ok {
			ok = append(ok, r)
		}
	}
	sort.Slice(ok, func(i, j int) bool { return ok[i].latency < ok[j].latency })

	fmt.Println("\n测速结果(前 10):")
	for i, r := range ok {
		if i >= 10 {
			break
		}
		fmt.Printf("  %-18s  %v\n", r.host, r.latency)
	}

	if len(ok) > n {
		ok = ok[:n]
	}
	var picked []string
	for _, r := range ok {
		picked = append(picked, r.host)
	}
	fmt.Printf("\n已选择最快的 %d 个: %v\n", len(picked), picked)
	return picked
}
