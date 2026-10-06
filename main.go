package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func main() {
	for {
		clearScreen()
		printBanner()
		fmt.Println()
		fmt.Println("   [1]  SSH Checker")
		fmt.Println("   [2]  SSH Cracker")
		fmt.Println("   [3]  Exit")
		fmt.Println()
		fmt.Print("Select option [1-3]: ")

		choice := askString()
		switch choice {
		case "1":
			runChecker()
		case "2":
			runCracker()
		case "3":
			fmt.Println("\nbye baby.")
			return
		default:
			fmt.Println("\ninvalid choice.")
			pause()
		}
	}
}

func askString() string {
	s, _ := reader.ReadString('\n')
	return strings.TrimSpace(s)
}

func askInt(defaultVal int) int {
	s := askString()
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}

func pause() {
	fmt.Print("\npress Enter to continue...")
	reader.ReadString('\n')
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func printBanner() {
	fmt.Println(colorCyan + `╔══════════════════════════════════════════════════════════════╗`)
	fmt.Println(`║              SSH TOOLKIT v1.0  (Go Edition)                  ║`)
	fmt.Println(`║              ─────────────────────────────                   ║`)
	fmt.Printf(`║              %-48s ║`+"\n", timeNow())
	fmt.Println(`╚══════════════════════════════════════════════════════════════╝` + colorReset)
}
