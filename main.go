package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/hpcloud/tail"
)

type CaddyLog struct {
	Timestamp float64 `json:"ts"`
	Status    int     `json:"status"`
	Duration  float64 `json:"duration"`
	Request   struct {
		RemoteIP string `json:"remote_ip"`
		Method   string `json:"method"`
		Host     string `json:"host"`
		URI      string `json:"uri"`
	} `json:"request"`
}

var (
	isTerminal      bool
	totalLines      int
	lastSize        int64
	startTime       time.Time
	lastH, lastW    int
	lastCPUTime     int64
	lastSampleTime  time.Time
	cachedLogs      []CaddyLog
	cachedStats     LogStats
)

type LogStats struct {
	RPS        float64
	Status1xx  float64
	Status2xx  float64
	Status3xx  float64
	Status4xx  float64
	Status5xx  float64
	TopIP      string
	TopIPPct   int
	AvgLatency int
}

func getProcessCPUTime() int64 {
	var usage syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &usage)
	return usage.Utime.Nano() + usage.Stime.Nano()
}

func calculateStats(logs []CaddyLog) LogStats {
	if len(logs) == 0 { return LogStats{} }
	var s LogStats
	ipMap := make(map[string]int)
	var totalLatency float64
	var count1, count2, count3, count4, count5 int
	minTs, maxTs := logs[0].Timestamp, logs[0].Timestamp

	for _, l := range logs {
		if l.Timestamp < minTs { minTs = l.Timestamp }
		if l.Timestamp > maxTs { maxTs = l.Timestamp }
		totalLatency += l.Duration
		ipMap[l.Request.RemoteIP]++
		switch {
		case l.Status >= 500: count5++
		case l.Status >= 400: count4++
		case l.Status >= 300: count3++
		case l.Status >= 200: count2++
		case l.Status >= 100: count1++
		}
	}

	total := float64(len(logs))
	s.Status1xx = (float64(count1) / total) * 100
	s.Status2xx = (float64(count2) / total) * 100
	s.Status3xx = (float64(count3) / total) * 100
	s.Status4xx = (float64(count4) / total) * 100
	s.Status5xx = (float64(count5) / total) * 100
	s.AvgLatency = int((totalLatency / total) * 1000)

	timeDiff := maxTs - minTs
	if timeDiff > 0 { s.RPS = total / timeDiff }

	topIP, topCount := "", 0
	for ip, count := range ipMap {
		if count > topCount {
			topIP = ip
			topCount = count
		}
	}
	s.TopIP = topIP
	s.TopIPPct = int((float64(topCount) / total) * 100)
	return s
}

func getTermSize() (int, int) {
	type winsize struct{ Row, Col, X, Y uint16 }
	ws := &winsize{}
	syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdin), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	return int(ws.Row), int(ws.Col)
}

func doubleClear() { fmt.Print("\033[H\033[2J\033[3J") }

func colorLabel(label string, active bool) string {
	if !isTerminal { return label }
	if active { return "\033[32m" + label + "\033[0m" }
	return "\033[31m" + label + "\033[0m"
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit { return fmt.Sprintf("%d B", b) }
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func countTotalLines(filePath string) (int, int64) {
	f, err := os.Open(filePath)
	if err != nil { return 0, 0 }
	defer f.Close()
	info, _ := f.Stat()
	count, buf := 0, make([]byte, 32*1024)
	for {
		c, err := f.Read(buf)
		if c > 0 { count += bytes.Count(buf[:c], []byte{'\n'}) }
		if err == io.EOF { break }
	}
	return count, info.Size()
}

func updateLineCount(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil { return false }
	currentSize := info.Size()
	if currentSize == lastSize { return false }
	if currentSize > lastSize {
		f, _ := os.Open(filePath)
		defer f.Close()
		f.Seek(lastSize, io.SeekStart)
		buf, bytesToRead := make([]byte, 32*1024), currentSize-lastSize
		for bytesToRead > 0 {
			readSize := int64(len(buf))
			if bytesToRead < readSize { readSize = bytesToRead }
			c, err := f.Read(buf[:readSize])
			if c > 0 {
				totalLines += bytes.Count(buf[:c], []byte{'\n'})
				bytesToRead -= int64(c)
			}
			if err != nil { break }
		}
		lastSize = currentSize
		return true
	}
	totalLines, lastSize = countTotalLines(filePath)
	return true
}

func getLastLines(filePath string, n int) ([]string, int64) {
	file, err := os.Open(filePath)
	if err != nil { return nil, 0 }
	defer file.Close()
	info, _ := file.Stat()
	fileSize := info.Size()
	var lines []string
	var cursor int64 = 0
	bufSize := int64(4096)
	if fileSize < bufSize { bufSize = fileSize }
	for cursor < fileSize {
		cursor += bufSize
		if cursor > fileSize { cursor = fileSize }
		file.Seek(fileSize-cursor, io.SeekStart)
		buf := make([]byte, bufSize)
		file.Read(buf)
		chunk := strings.Split(string(buf), "\n")
		if len(lines) == 0 && chunk[len(chunk)-1] == "" { chunk = chunk[:len(chunk)-1] }
		lines = append(chunk, lines...)
		if len(lines) > n { return lines[len(lines)-n:], fileSize }
	}
	return lines, fileSize
}

func isAsset(uri string) bool {
	clean := strings.Split(uri, "?")[0]
	assets := map[string]bool{".js": true, ".css": true, ".map": true, ".scss": true, ".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".avif": true, ".svg": true, ".ico": true, ".cur": true, ".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true, ".mp4": true, ".webm": true, ".mov": true, ".ogv": true, ".mp3": true, ".wav": true, ".m4a": true, ".ogg": true, ".flac": true, ".aac": true, ".zip": true, ".gz": true, ".tar": true, ".pdf": true, ".webmanifest": true, ".xml": true, ".robots.txt": true, ".php": true}
	lastDot := strings.LastIndex(clean, ".")
	return lastDot != -1 && assets[strings.ToLower(clean[lastDot:])]
}

func drawStatusBar(w int, dash bool) {
	if !isTerminal { return }
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	now := time.Now()
	currentCPUTime := getProcessCPUTime()
	cpuDelta := currentCPUTime - lastCPUTime
	timeDelta := now.Sub(lastSampleTime).Nanoseconds()
	cpuPercent := 0.0
	if timeDelta > 0 { cpuPercent = (float64(cpuDelta) / float64(timeDelta)) * 100 }
	lastCPUTime, lastSampleTime = currentCPUTime, now
	uptime := time.Since(startTime).Round(time.Second)
	fmt.Fprintf(os.Stderr, "%s\033[K\n", strings.Repeat("━", w))
	fmt.Fprintf(os.Stderr, "\033[1;37mSystem Stats:\033[0m Uptime: %-7s | Mem: %-8s | CPU: %-4.1f%% | Dashboard: %s\033[K\n",
		uptime, formatBytes(int64(m.Alloc)), cpuPercent, colorLabel("Active", dash))
}

func main() {
	startTime = time.Now()
	lastSampleTime = time.Now()
	lastCPUTime = getProcessCPUTime()
	var lCount int
	var ha, e, all, clear, count, status, dash bool
	var f, host string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\033[1mCLOG v0.6.3 - High-Visibility Caddy Logs\033[0m\n")
		fmt.Fprintf(os.Stderr, "Usage: clog [options] <logfile>\n\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --lines, -l         number of previous lines to show\n")
		fmt.Fprintf(os.Stderr, "  --host, -h          only show logs for this domain or IP\n")
		fmt.Fprintf(os.Stderr, "  --find, -f          only show lines containing this string\n")
		fmt.Fprintf(os.Stderr, "  --errors, -e        only show requests with status >= 400\n")
		fmt.Fprintf(os.Stderr, "  --hide-assets, -ha  hides common asset types (.js, .css, images, etc)\n")
		fmt.Fprintf(os.Stderr, "  --all, -a           show entire history and ignore asset filters\n")
		fmt.Fprintf(os.Stderr, "  --status, -s        show system resource bar at bottom\n")
		fmt.Fprintf(os.Stderr, "  --dashboard, -d     enable 1-second dashboard mode\n")
		fmt.Fprintf(os.Stderr, "  --clear-screen, -c  clear terminal before starting and on exit\n")
		fmt.Fprintf(os.Stderr, "  --help              show this help menu\n\n")
	}

	flag.IntVar(&lCount, "l", 0, "")
	flag.IntVar(&lCount, "lines", 0, "")
	flag.BoolVar(&ha, "ha", false, "")
	flag.BoolVar(&ha, "hide-assets", false, "")
	flag.BoolVar(&e, "e", false, "")
	flag.BoolVar(&e, "errors", false, "")
	flag.StringVar(&f, "f", "", "")
	flag.StringVar(&f, "find", "", "")
	flag.StringVar(&host, "h", "", "")
	flag.StringVar(&host, "host", "", "")
	flag.BoolVar(&all, "a", false, "")
	flag.BoolVar(&all, "all", false, "")
	flag.BoolVar(&clear, "c", false, "")
	flag.BoolVar(&clear, "clear-screen", false, "")
	flag.BoolVar(&count, "co", false, "")
	flag.BoolVar(&count, "count", false, "")
	flag.BoolVar(&status, "s", false, "")
	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&dash, "d", false, "")
	flag.BoolVar(&dash, "dashboard", false, "")
	flag.Parse()

	if flag.NArg() < 1 { flag.Usage(); os.Exit(1) }
	filePath := flag.Arg(0)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "\033[31mError:\033[0m Log file '%s' not found.\n", filePath)
		os.Exit(1)
	}

	fileInfo, _ := os.Stdout.Stat()
	isTerminal = (fileInfo.Mode() & os.ModeCharDevice) != 0
	totalLines, lastSize = countTotalLines(filePath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if isTerminal {
			fmt.Print("\033[?25h")
			if clear || dash { doubleClear() }
		}
		os.Exit(0)
	}()

	if isTerminal {
		if dash { fmt.Print("\033[?25l") }
		if clear || dash { doubleClear() }
	}

	firstRun := true
	for {
		hasChanged := updateLineCount(filePath)
		h, w := getTermSize()
		if isTerminal && dash && (h != lastH || w != lastW) {
			doubleClear()
			lastH, lastW = h, w
			hasChanged = true
		}
		headerHeight := 7
		if host != "" || f != "" { headerHeight = 8 }
		if status { headerHeight += 2 }

		currentLCount := lCount
		if dash && currentLCount == 0 {
			currentLCount = h - headerHeight
			if currentLCount < 1 { currentLCount = 1 }
		} else if currentLCount == 0 && !all {
			currentLCount = 10
		}

		if isTerminal && dash { fmt.Print("\033[H") }
		processLogs(filePath, currentLCount, ha, e, all, f, host, count, dash, w, hasChanged || firstRun)
		firstRun = false

		if isTerminal && dash && status { drawStatusBar(w, dash) }
		if !dash { break }
		if isTerminal { fmt.Print("\033[J") }
		time.Sleep(1 * time.Second)
	}
}

func processLogs(filePath string, lCount int, ha bool, e bool, all bool, f string, host string, count bool, dash bool, width int, shouldUpdate bool) {
	if shouldUpdate {
		fetchCount := lCount * 10
		if all { fetchCount = totalLines }
		if fetchCount < 10 { fetchCount = 10 }
		rawLines, _ := getLastLines(filePath, fetchCount)
		newLogs := make([]CaddyLog, 0, len(rawLines))
		for _, line := range rawLines {
			if line == "" { continue }
			var l CaddyLog
			if err := json.Unmarshal([]byte(line), &l); err == nil { newLogs = append(newLogs, l) }
		}
		cachedLogs = newLogs
		cachedStats = calculateStats(cachedLogs)
	}

	if isTerminal || count {
		fmt.Fprintf(os.Stderr, "%s\033[K\n", strings.Repeat("━", width))
		fmt.Fprintf(os.Stderr, "\033[1;37mWatching:\033[0m %s\033[K\n", filePath)
		numLabel := colorLabel(fmt.Sprintf("Lines %d", lCount), lCount > 0)
		if all { numLabel = colorLabel("All History", true) }
		fmt.Fprintf(os.Stderr, "\033[1;37mActive Flags:\033[0m %s | %s | %s | %s | %s | %s\033[K\n", numLabel, colorLabel("Host", host != ""), colorLabel("Find", f != ""), colorLabel("Errors", e), colorLabel("Hide Assets", ha), colorLabel("Dashboard", dash))
		fmt.Fprintf(os.Stderr, "\033[1;37mLog Stats:\033[0m RPS: %.1f | 1xx: %.0f%% | 2xx: %.0f%% | 3xx: %.0f%% | 4xx: %.0f%% | 5xx: %.0f%% | Latency: %dms | Top IP: %s (%d%%)\033[K\n", cachedStats.RPS, cachedStats.Status1xx, cachedStats.Status2xx, cachedStats.Status3xx, cachedStats.Status4xx, cachedStats.Status5xx, cachedStats.AvgLatency, cachedStats.TopIP, cachedStats.TopIPPct)
		fmt.Fprintf(os.Stderr, "\033[1;37mFile Stats:\033[0m %d lines | %s\033[K\n", totalLines, formatBytes(lastSize))
		if host != "" || f != "" {
			details := []string{}
			if host != "" { details = append(details, "Host: "+host) }
			if f != "" { details = append(details, "Find: "+f) }
			fmt.Fprintf(os.Stderr, "\033[1;37mFilters:\033[0m %s\033[K\n", strings.Join(details, " | "))
		}
		fmt.Fprintf(os.Stderr, "%s\033[K\n", strings.Repeat("━", width))
	}

	var finalOutput []string
	processedCount := 0
	for i := len(cachedLogs) - 1; i >= 0; i-- {
		l := cachedLogs[i]
		if host != "" {
			h := strings.ToLower(host)
			if !strings.Contains(strings.ToLower(l.Request.RemoteIP), h) && !strings.Contains(strings.ToLower(l.Request.Host), h) { continue }
		}
		if !all {
			if ha && isAsset(l.Request.URI) { continue }
			if processedCount >= lCount { break }
		}
		if e && l.Status < 400 { continue }
		if f != "" {
			fLow := strings.ToLower(f)
			if !strings.Contains(strings.ToLower(l.Request.URI), fLow) && !strings.Contains(strings.ToLower(l.Request.Host), fLow) && !strings.Contains(strings.ToLower(l.Request.RemoteIP), fLow) { continue }
		}
		tm := time.Unix(int64(l.Timestamp), 0).Format("2006-01-02 | 15:04:05")
		st := fmt.Sprintf("%d", l.Status)
		if isTerminal {
			color := "\u001b[32m"
			if l.Status >= 400 { color = "\u001b[31m" }
			st = color + st + "\u001b[0m"
		}
		out := fmt.Sprintf("%s | %-15s | %s | %-6s | %4dms | %s%s", tm, l.Request.RemoteIP, st, l.Request.Method, int(l.Duration*1000), l.Request.Host, l.Request.URI)
		if dash && len(out) > width { out = out[:width-3] + "..." }
		finalOutput = append([]string{out}, finalOutput...)
		processedCount++
	}
	for _, line := range finalOutput { fmt.Printf("%s\033[K\n", line) }

	if !dash {
		voidLogger := log.New(io.Discard, "", 0)
		config := tail.Config{Follow: true, MustExist: true, Logger: voidLogger, Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}}
		t, _ := tail.TailFile(filePath, config)
		for line := range t.Lines {
			var l CaddyLog
			if err := json.Unmarshal([]byte(line.Text), &l); err != nil { continue }
			if host != "" {
				h := strings.ToLower(host)
				if !strings.Contains(strings.ToLower(l.Request.RemoteIP), h) && !strings.Contains(strings.ToLower(l.Request.Host), h) { continue }
			}
			if ha && isAsset(l.Request.URI) { continue }
			if e && l.Status < 400 { continue }
			if f != "" {
				fLow := strings.ToLower(f)
				if !strings.Contains(strings.ToLower(l.Request.URI), fLow) && !strings.Contains(strings.ToLower(l.Request.Host), fLow) && !strings.Contains(strings.ToLower(l.Request.RemoteIP), fLow) { continue }
			}
			tm := time.Unix(int64(l.Timestamp), 0).Format("2006-01-02 | 15:04:05")
			st := fmt.Sprintf("%d", l.Status)
			if isTerminal {
				color := "\u001b[32m"
				if l.Status >= 400 { color = "\u001b[31m" }
				st = color + st + "\u001b[0m"
			}
			fmt.Printf("%s | %-15s | %s | %-6s | %4dms | %s%s\033[K\n", tm, l.Request.RemoteIP, st, l.Request.Method, int(l.Duration*1000), l.Request.Host, l.Request.URI)
		}
	}
}
