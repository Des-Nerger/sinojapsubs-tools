package main

import (
	"bufio"
	"io"
	"os"
	"unicode"
)

func main() {
	left, right := "", ""
	switch len(os.Args) {
	case 3:
		right = os.Args[2]
		fallthrough
	case 2:
		left = os.Args[1]
	default:
		os.Stderr.WriteString("!(2 <= len(os.Args) <= 3)")
		return
	}
	br, bw := bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout); defer bw.Flush()
	prevWasHan := false
	for {
		r, _, err := br.ReadRune()
		switch err {
		case nil:
		case io.EOF: return
		default: panic(err)
		}
		prevWasHan = func() bool {
			if unicode.Is(unicode.Han, r) {
				if !prevWasHan {bw.WriteString(left)}
				return true
			}
			if prevWasHan {bw.WriteString(right)}
			return false
		} ()
		bw.WriteRune(r)
	}
}
