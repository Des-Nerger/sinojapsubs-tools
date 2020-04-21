package main
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	type entry struct {
		lineNumber int
		hanzis, definition string
	}
	containHanzis := func(entries []entry, hanzis string) bool {
		for _, e := range entries {
			if e.hanzis == hanzis {return true}
		}
		return false
	}

	d := map[string][]entry{}
	lastFirstFileLineNumber := -1
	{
		//hanzisAlreadyOccured := map[string]struct{}{}
		lineNumber := 0
		for i, arg := range os.Args[1:] {
			func() {
				file, err := os.Open(arg)
				if err != nil {
					panic(err)
				}
				defer file.Close()
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					lineNumber++
					s := strings.Split(scanner.Text(), "\t")
					switch len(s) {
					case 0, 1:
						panic(fmt.Sprintf("len(%#v)<=1", s))
					case 2:
						s = append(s, "")
					}
				/*
					if _, ok := hanzisAlreadyOccured[s[0]]; ok {
						if s[len(s)-1]!="" {fmt.Printf("skipping %v\n", s)}
						continue
					}
					hanzisAlreadyOccured[s[0]] = struct{}{}
				*/
					e := entry{lineNumber, s[0], s[len(s)-1]}
					for _, p := range s[1:len(s)-1] {
						if containHanzis(d[p], e.hanzis) {continue}
						d[p] = append(d[p], e)
					}
				}
				//fmt.Println(lineNumber)
				if i==0 {
					lastFirstFileLineNumber = lineNumber
				}
			} ()
		}
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line=="" {continue}
		printMore := false
		if line[len(line)-1]=='+' {
			line = line[:len(line)-1]
			printMore = true
		}
		entries := d[strings.TrimSpace(line)]
		for {
			foundInFirstFile := false
			for _, e := range entries {
				if !printMore {
					if e.lineNumber>lastFirstFileLineNumber {break}
					foundInFirstFile = true
				}
				if e.lineNumber>lastFirstFileLineNumber {
					fmt.Printf("%v ", e.hanzis)
				} else { 
					fmt.Printf("%7v  %-10v  %v\n", e.lineNumber, e.hanzis, e.definition)
				}
			}
			if printMore {
				fmt.Println()
				break
			}
			if foundInFirstFile {break}
			printMore = true
		}
	}
}
