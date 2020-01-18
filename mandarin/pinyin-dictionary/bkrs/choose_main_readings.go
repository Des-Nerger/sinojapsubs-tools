package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

/*
type MainReadingChooser map[rune]string
func (c *MainReadingChooser) Choose([])
*/

func main() {
	hanziMainReadings := map[rune]string{}
	func() {
		file, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := strings.Split(scanner.Text(), "\t")
			runes := []rune(s[0])
			if len(s)!=2 || len(runes)!=1 {panic(fmt.Sprintln(s))}
			hanziMainReadings[runes[0]] = s[1]
		}
	} ()
	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
outermostLoop:
	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, "\t")
		if len(s) <= 2 {
			fmt.Fprintln(bw, line)
			continue
		}
		chosenSyllables := []string(nil)
		for i, reading := range s[1:] {
			syllables := strings.Split(reading, " ")
			if i==0 {chosenSyllables = syllables; continue}
			j:=0
			for _, hanzi := range s[0] {
				if chosenSyllables[j] != syllables[j] {
					hanziMainReading := hanziMainReadings[hanzi]
					switch hanziMainReading {
					case chosenSyllables[j], syllables[j]:
					default:
						fmt.Fprintf(os.Stderr, "%v: %c: \"%v\" vs \"%v\", while main reading is \"%v\"; skipping\n",
							s[0], hanzi, chosenSyllables[j], syllables[j], hanziMainReading,
						)
						continue outermostLoop
					}
					chosenSyllables[j] = hanziMainReading
				}
				j++
			}
		}
		bw.WriteString(s[0])
		bw.WriteByte('\t')
		bw.WriteString(strings.Join(chosenSyllables, " "))
		bw.WriteByte('\n')
	}
}
