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
		switch r {
		//case 'ー', '～', '・', '･':
		case '…':
			bw.WriteString("...")
		case '！':
			bw.WriteByte('!')
		case '？':
			bw.WriteByte('?')
		case '～':
			bw.WriteByte('~')
		case '・', '･':
			bw.WriteString(blankSequence)
		default:
			if unicode.In(r, unicode.Pi, unicode.Ps) {
				bw.WriteByte('[')
			} else if unicode.In(r, unicode.Pf, unicode.Pe) {
				bw.WriteByte(']')
			} else if !unicode.In(r,
				unicode.Hiragana,
				unicode.Katakana,
				unicode.Han,
				FullwidthAlphaNum,
				/*
				unicode.Latin,
				unicode.Punct,
				unicode.Symbol,
				*/
				unicode.Lm,
				unicode.Bopomofo,
				unicode.Sk,
			) {
				bw.WriteRune(r)
			} else {
				bw.WriteString(blankSequence)
			}
		}
	}
}
