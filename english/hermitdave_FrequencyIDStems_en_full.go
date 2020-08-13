package main
import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	. "unsafe"

	"github.com/Des-Nerger/porter2"
)
func main() {
	panicCheck := func(e error) {if e!=nil {panic(e)}}
	sign := func(i int) int {isntZero:=i!=0; return i>>(strconv.IntSize-1) | int(*(*byte)(Pointer(&isntZero)))}
	m := make(map[string]int)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s := strings.SplitN(scanner.Text(), " ", 2)
		count, e := strconv.Atoi(s[1]); panicCheck(e)
		idStart := -1
	innerFor:
		for i:=0;; {
			var (r rune; size int)
			switch sign(i-len(s[0])) {
			case -1: r,size=utf8.DecodeRuneInString(s[0][i:]); if r==utf8.RuneError{panic("failed to decode rune")}
			case  0: size=1
			case +1: break innerFor
			}
			if idStart==-1 {
				if unicode.Is(unicode.Letter, r) {idStart=i}
			} else {
				if !unicode.In(r, unicode.Letter, unicode.Digit) {
					m[porter2.Stem(s[0][idStart:i])] += count
					idStart=-1
				}
			}
			i += size
		}
	}
	bw:=bufio.NewWriter(os.Stdout); defer bw.Flush()
	for id, count := range m {
		bw.WriteString(id); bw.WriteByte(' '); bw.WriteString(strconv.Itoa(count)); bw.WriteByte('\n')
	}
}
