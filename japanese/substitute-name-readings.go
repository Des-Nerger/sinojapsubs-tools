package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/miiton/kanaconv"
)

func main() {
	kanjiNamesDict := map[[2]string]struct{}{}
	func () {
		file, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		sc := bufio.NewScanner(file)
	outerLoop:
		for sc.Scan() {
			splits := strings.Split(sc.Text(), "\t")
			if len(splits)<=1 {continue}
			for _, r := range splits[0] {
				if !unicode.In(r, unicode.Han) {
					continue outerLoop
				}
			}
			kanjiNamesDict[[...]string{splits[0], splits[1]}] = struct{}{}
		}
	} ()
	

	j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
	j.Start()
	defer j.Wait()

	bw := bufio.NewWriter(os.Stdout)
	defer bw.Flush()
	lastUsedReadings := map[string]string{}
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := sc.Text()
		inParentheses := false
		currentSubstringStart := 0
		type reading struct {int; string}
		readings := []reading(nil)
		keptLineParts := []string(nil)
		droppedRunesCount := 0
		for i, r := range line {
			if inParentheses {
				if r == ')' {
					readings = append(readings,
						reading{currentSubstringStart-droppedRunesCount, line[currentSubstringStart+1:i]},
					)
					droppedRunesCount += i-currentSubstringStart + 1
					currentSubstringStart = i+1
					inParentheses = false
				}
			} else {
				if r == '(' {
					keptLineParts = append(keptLineParts, line[currentSubstringStart:i])
					currentSubstringStart = i
					inParentheses = true
				}
			}
		}
		switch {
		case inParentheses:
			fmt.Fprintf(os.Stderr, "unterminated '(' on line: «%v»; ", line)
			//fmt.Fprintln(os.Stderr, "ignoring the unterminated part")
			fmt.Fprintln(os.Stderr, "leaving the unterminated part untouched"); fallthrough
		case currentSubstringStart < len(line):
			keptLineParts = append(keptLineParts, line[currentSubstringStart:])
		}
		//fmt.Fprintf(bw, "%#v\n%v\n\n", strings.Join(keptLineParts, ""), readings)
		morphemes := j.AnalyzeLine(strings.Join(keptLineParts, ""))
		curPos := 0
		for _, morpheme := range morphemes {
			nextPos := curPos + len(morpheme[0])
			specifiedReading := ""
			for _, reading := range readings {
				if reading.int == nextPos {
					specifiedReading = reading.string
					break
				}
			}
			bw.WriteString(func() string {
				if specifiedReading == "" {
					lastUsedReading := lastUsedReadings[morpheme[0]]
					if lastUsedReading == "" {
						return morpheme[0]
					}
					if utf8.RuneCountInString(morpheme[0]) == 1 {
						fmt.Fprintf(os.Stderr, "«%v»: !!! using last specified reading: %v|%v !!!\n",
							line, morpheme[0], lastUsedReading,
						)
					} else {
						fmt.Fprintf(os.Stderr, "«%v»: using last specified reading: %v|%v\n", line, morpheme[0], lastUsedReading)
					}
					return lastUsedReading
				}
				if _, ok := kanjiNamesDict[[...]string{morpheme[0], specifiedReading}]; ok {
					katakanaReading := kanaconv.HiraganaToKatakana(specifiedReading)
					lastUsedReadings[morpheme[0]] = katakanaReading
					return katakanaReading
				}
				return morpheme[0] + "(" + specifiedReading + ")"
			} ())
			curPos = nextPos
		}
		bw.WriteByte('\n')
	}
}
