package main
import (
	"bufio"
	//"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/hermanschaaf/cedict"
	"golang.org/x/text/unicode/norm"
)

func main() {
	c := cedict.New(os.Stdin)
	wc := norm.NFC.Writer(os.Stdout)
	bw := bufio.NewWriter(wc)
	defer func() {
		bw.Flush()
		wc.Close()
	} ()
	for {
		err := c.NextEntry()
		if err != nil {
			break
		}
		entry := c.Entry()
		if utf8.RuneCountInString(entry.Simplified)<=1 || !containsOnly(entry.Simplified, unicode.Han) {continue}
		//fmt.Fprintf(os.Stderr, "%#v %#v\n", entry.Simplified, entry.Pinyin)
		bw.WriteString(entry.Simplified)
		bw.WriteByte('\t')
		bw.WriteString(pinyin( strings.ToLower(entry.Pinyin) ))
		bw.WriteByte('\n')
	}
}

/*
func contains(s string, ranges ...*unicode.RangeTable) bool {
	for _, r := range s {
		if unicode.In(r, ranges...) {return true}
	}
	return false
}
*/

func containsOnly(s string, ranges ...*unicode.RangeTable) bool {
	for _, r := range s {
		if !unicode.In(r, ranges...) {return false}
	}
	return true
}
