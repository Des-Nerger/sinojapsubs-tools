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
		curIsHan := unicode.Is(unicode.Han, r)
		if curIsHan {
			if !prevWasHan {bw.WriteString(left)}
		} else {
			if prevWasHan {bw.WriteString(right)}
		}
		bw.WriteRune(r)
		prevWasHan = curIsHan
	}
}
