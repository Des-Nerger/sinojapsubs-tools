package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
)

func panicIfNotNil(e error) {if e != nil {panic(e)}}

func main() {
	freq := map[rune]int{}
	i := -1
	for sc:=bufio.NewScanner(os.Stdin); sc.Scan(); {
		func() {
			file, e := os.Open(sc.Text()); panicIfNotNil(e); defer file.Close()
			i++; if i%128==0 {fmt.Fprintln(os.Stderr, i)}
			br := bufio.NewReader(file)
		loop:
			for {
				r, _, e := br.ReadRune()
				switch e {
				case nil:
				case io.EOF: break loop
				default: panic(e)
				}
				if unicode.Is(unicode.Han, r) {freq[r]++}
			}
		} ()
	}
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	for r,n := range freq {
		bw.WriteString(strconv.Itoa(n))
		bw.WriteByte('\t')
		bw.WriteRune(r)
		bw.WriteByte('\n')
	}
}
