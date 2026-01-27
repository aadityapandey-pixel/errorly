package aierror

import (
	"fmt"
	"strings"
)

func colorRed(s string) string    { return "\033[31m" + s + "\033[0m" }
func colorYellow(s string) string { return "\033[33m" + s + "\033[0m" }
func colorCyan(s string) string   { return "\033[36m" + s + "\033[0m" }

func PrintFormatted(aiResponse string) {
	fmt.Println("\n==================================================")
	fmt.Println(colorRed("🔴 AI ERROR ANALYSIS"))
	fmt.Println("==================================================")

	lines := strings.Split(aiResponse, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "ROOT CAUSE:"):
			fmt.Println(colorRed(line))
		case strings.HasPrefix(line, "WHY IT HAPPENED:"):
			fmt.Println(colorYellow(line))
		case strings.HasPrefix(line, "HOW TO FIX:"):
			fmt.Println(colorCyan(line))
		case strings.HasPrefix(line, "EXAMPLE FIX CODE:"):
			fmt.Println(colorCyan(line))
		default:
			fmt.Println("  " + line)
		}
	}

	fmt.Println("==================================================\n")
}
