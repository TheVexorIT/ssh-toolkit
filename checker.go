package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type checkResult struct {
	Target string
	Banner string
	Open   bool
}

func runChecker() {
	clearScreen()
	printBox("SSH CHECKER")

	fmt.Print("\nIP list (drag file or paste path): ")
	ipPath := askString()
	if ipPath == "" {
		fail("no path given")
		pause()
		return
	}
	ipPath = cleanPath(ipPath)

	ips, err := readLines(ipPath)
	if err != nil {
		fail("can't read file: " + err.Error())
		pause()
		return
	}

	fmt.Printf("\nLoaded %d targets.\n", len(ips))

	fmt.Print("Threads [default 100]: ")
	threads := askInt(100)
	if threads < 1 {
		threads = 100
	}

	fmt.Print("Timeout seconds [default 5]: ")
	timeout := askInt(5)
	if timeout < 1 {
		timeout = 5
	}

	// make results dir
	checkerDir := filepath.Join("results", "checker")
	os.MkdirAll(checkerDir, 0755)

	goodsPath := filepath.Join(checkerDir, "goods.txt")
	badsPath := filepath.Join(checkerDir, "bads.txt")

	goodsFile, _ := os.Create(goodsPath)
	badsFile, _ := os.Create(badsPath)
	defer goodsFile.Close()
	defer badsFile.Close()

	gw := bufio.NewWriter(goodsFile)
	bw := bufio.NewWriter(badsFile)
	defer gw.Flush()
	defer bw.Flush()

	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, threads)
		mu      sync.Mutex
		done    int
		goodCnt int
		badCnt  int
	)

	start := time.Now()
	info(fmt.Sprintf("Starting check with %d threads, timeout %ds...", threads, timeout))

	for _, target := range ips {
		wg.Add(1)
		sem <- struct{}{}
		go func(t string) {
			defer wg.Done()
			defer func() { <-sem }()

			host, port := normalizeTarget(t, "22")
			res := checkSSH(host, port, time.Duration(timeout)*time.Second)

			mu.Lock()
			defer mu.Unlock()
			done++
			if res.Open {
				goodCnt++
				fmt.Fprintf(gw, "%s | %s\n", t, res.Banner)
				success(fmt.Sprintf("[%d/%d] %s  %s", done, len(ips), t, res.Banner))
			} else {
				badCnt++
				fmt.Fprintf(bw, "%s\n", t)
			}
			if done%50 == 0 {
				gw.Flush()
				bw.Flush()
			}
		}(target)
	}

	wg.Wait()
	gw.Flush()
	bw.Flush()

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Println(colorGreen + "═══════════════════════════════════════════════════" + colorReset)
	fmt.Printf("  Duration     : %.1fs\n", elapsed.Seconds())
	fmt.Printf("  Total        : %d\n", len(ips))
	fmt.Printf("  Good (SSH)   : %d\n", goodCnt)
	fmt.Printf("  Bad          : %d\n", badCnt)
	fmt.Println(colorGreen + "═══════════════════════════════════════════════════" + colorReset)
	fmt.Println("  results/checker/goods.txt")
	fmt.Println("  results/checker/bads.txt")
	pause()
}

func checkSSH(host, port string, timeout time.Duration) checkResult {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return checkResult{Target: addr, Open: false}
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return checkResult{Target: addr, Open: false}
	}

	banner := strings.TrimSpace(string(buf[:n]))
	ok := strings.HasPrefix(banner, "SSH-")
	return checkResult{Target: addr, Banner: banner, Open: ok}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		lines = append(lines, l)
	}
	return lines, sc.Err()
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.Trim(p, `"'`)
	p = strings.TrimPrefix(p, "& ")
	p = strings.TrimPrefix(p, "file:///")
	return p
}

func normalizeTarget(t, defaultPort string) (string, string) {
	t = strings.TrimSpace(t)
	t = strings.TrimPrefix(t, "ssh://")
	if strings.Contains(t, ":") {
		parts := strings.SplitN(t, ":", 2)
		return parts[0], parts[1]
	}
	return t, defaultPort
}
