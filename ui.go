package main

import (
	"fmt"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[91m"
	colorGreen  = "\033[92m"
	colorYellow = "\033[93m"
	colorBlue   = "\033[94m"
	colorCyan   = "\033[96m"
	colorWhite  = "\033[97m"
	colorBold   = "\033[1m"
)

func timeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func success(msg string) {
	fmt.Println(colorGreen + "[✓] " + msg + colorReset)
}

func fail(msg string) {
	fmt.Println(colorRed + "[✗] " + msg + colorReset)
}

func info(msg string) {
	fmt.Println(colorCyan + "[*] " + msg + colorReset)
}

func warn(msg string) {
	fmt.Println(colorYellow + "[!] " + msg + colorReset)
}

func printBox(title string) {
	fmt.Println(colorBlue + "╔══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ %-60s ║\n", title)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝" + colorReset)
}
