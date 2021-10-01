package utils

import (
	"bufio"
	"fmt"
	"os"
)

func Pause() {
	fmt.Printf("\nPress Enter to continue ... ")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Find(s []string, e string) int {
	for i, item := range s {
		if item == e {
			return i
		}
	}
	return -1
}

func Prepend(s []string, e string) []string {
	s = append(s, e)
	copy(s[1:], s)
	s[0] = e
	return s
}

func HasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func Flush() {
	fmt.Print("\033[H\033[2J")
}
