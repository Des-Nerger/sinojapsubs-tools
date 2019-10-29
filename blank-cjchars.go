package main

import (
	"bufio"
	"io"
	"os"
	"unicode"

	. "github.com/Des-Nerger/sinojapsubs-tools/commonrangetables"
)

func main() {
	blankSequence := "囗" //口◻⬜▢◯
	switch len(os.Args) {
	case 1:
	case 2:
		blankSequence = os.Args[1]
	default:
		os.Stderr.WriteString("!(1 <= len(os.Args) <= 2)")
		return
	}
	br, bw := bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout)
	for {
		r, _, err := br.ReadRune()
		switch err {
		case nil: // Do nothing
		case io.EOF:
			bw.Flush()
			return
		default:
			panic(err)
		}
		if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han, FullwidthAlphaNum) {
			bw.WriteString(blankSequence)
			continue
		}
		bw.WriteRune(r)
	}
}
