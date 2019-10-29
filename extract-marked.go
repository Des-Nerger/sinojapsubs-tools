package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	extractedStrings := []string{}
	extractedStringsSet := map[string]struct{}{}
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := sc.Text()
		markedStart := -1
		for i, r := range line {
			switch markedStart {
			case -1:
				switch r {
				case '<':
					markedStart = i+1
				}
			default:
				switch r {
				case '<':
					fmt.Fprintf(os.Stderr, "found nested '<', ignoring previous %q\n", line[markedStart-1:i])
					markedStart = i+1
				case '>':
					extractedString := line[markedStart:i]
					if _, ok := extractedStringsSet[extractedString]; !ok {
						extractedStringsSet[extractedString] = struct{}{}
						extractedStrings = append(extractedStrings, extractedString)
					}
					markedStart = -1
				case '|':
					if line[markedStart-1] == '|' {
						fmt.Fprintf(os.Stderr, "found another '|', ignoring previous %q\n", line[markedStart-1:i])
					}
					markedStart = i+1
				}
			}
		}
		if markedStart != -1 {
			fmt.Fprintf(os.Stderr, "unterminated marking, ignoring final %q\n", line[markedStart-1:])
		}
	}
	for _, extractedString := range extractedStrings {
		fmt.Printf("%v\t\n", extractedString)
	}
	fmt.Fprintf(os.Stderr, "len(extractedStrings) == %v\n", len(extractedStrings))
}
