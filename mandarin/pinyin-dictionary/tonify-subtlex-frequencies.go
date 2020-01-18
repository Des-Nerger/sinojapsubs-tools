package main
import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var debugMode = false

func main() {
	wc := norm.NFC.Writer(os.Stdout)
	bw := bufio.NewWriter(wc)
	defer func() {
		bw.Flush()
		wc.Close()
	} ()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		f := strings.Split(scanner.Text(), "\t")
		//fmt.Fprintf(os.Stderr, "%#v\n", f)
		if f[2]=="#" {continue}
		altSyllables := [][]string(nil)
		for _, s := range strings.Split(strings.ToLower(f[2]), " ") {
			altSyllable := strings.Split(s, "/")
			for i := range altSyllable {
				//work around «穆罕默德·布迪亚	8	mu4 han3 mo4 de2 #bu4 di2 ya4	muhanmode*budiya...» line:
				if altSyllable[i][0]=='#' {altSyllable[i] = altSyllable[i][1:]}

				altSyllable[i] = pinyin(altSyllable[i])
			}
			altSyllables = append(altSyllables, altSyllable)
		}
		bw.WriteString(f[0])
		if f[0]==/*"龙马会"*/ "龙虎榜" {
			debugMode = true
		}
		combinations := allCombinations(altSyllables)
		debugMode = false
		for _, combination := range combinations {
			bw.WriteByte('\t')
			bw.WriteString(strings.Join(combination, "'"))
		}
		bw.WriteByte('\t')
		bw.WriteString(f[14])
		bw.WriteByte('\n')
	}
}

func allCombinations(alternatives [][]string) (combinations [][]string) {
	switch len(alternatives) {
	case 1:
		for _, alternative := range alternatives[0] {
			combinations = append(combinations, []string{alternative})
		}
	/*
		fallthrough
	case 0:
	*/
		return
	}
	tailAlternatives := allCombinations(alternatives[1:])
	for _, headAlternative := range alternatives[0] {
		for _, tailAlternative := range tailAlternatives {
			combinations = append(combinations, append([]string{headAlternative}, tailAlternative...))
		}
	}
	if debugMode {
		fmt.Fprintf(os.Stderr, "%v --> %v\n", alternatives, combinations)
	}
	return
}
