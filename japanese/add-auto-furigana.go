package main
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Des-Nerger/sinojapsubs-tools/japanese/diff"
	"github.com/miiton/kanaconv"
	"github.com/shogo82148/go-mecab"
)

func main() {
	panicCheck := func(err error) {if err != nil {panic(err)}}
	tagger, err := mecab.New(map[string]string{"eos-format": "\x00"}); panicCheck(err); defer tagger.Destroy()

	scanner,bw := bufio.NewScanner(os.Stdin),bufio.NewWriter(os.Stdout); defer bw.Flush()
	handleLineFlush := func() func() {
		var lineBuffered bool
		flag.BoolVar(&lineBuffered, "lineBuffered", false, "")
		flag.Parse()
		return func() {if lineBuffered {bw.Flush()}}
	} ()
	blanked := make(map[string]struct{}, flag.NArg())
	for _, a := range flag.Args() {blanked[a] = struct{}{}}
	contains := func(s string, rt *unicode.RangeTable) bool {
		for _, r := range s {if unicode.Is(rt, r) {return true}}; return false
	}
	blanks, max := "", func(a, b int) int {if a>b {return a}; return b}
	for scanner.Scan() {
		line := scanner.Text()
		if !contains(line, unicode.Han) {fmt.Fprintln(bw, line); handleLineFlush(); continue}
		result, err := tagger.Parse(line); panicCheck(err)
		i := 0
		for _, word := range strings.FieldsFunc(result, func(r rune)bool{return r=='\n'}) {
			const sep = ",\t"
			if strings.HasPrefix(word, sep) {continue}
			fields := strings.FieldsFunc(word, func(r rune)bool{return strings.ContainsRune(sep, r)})
			writing, yomi := fields[0], kanaconv.KatakanaToHiragana(fields[len(fields)-2])
			writingStart := i + strings.Index(line[i:], writing)
			bw.WriteString(line[i:writingStart])
			i = writingStart + len(writing)
			switch _, ok := blanked[fields[len(fields)-3]]; {
			case ok:
				runeCount := utf8.RuneCountInString(writing)
				const blank = "　"/*"－"*/
				writtenBlanksLen := len(blank)*runeCount
				if len(blanks)<writtenBlanksLen {blanks=strings.Repeat(blank, max(32, 2*runeCount))}
				writing=blanks[:writtenBlanksLen]
				fallthrough
			case !contains(writing, unicode.Han):
				bw.WriteString(writing)
				continue
			}
			const format = /*"\uFFF9%v\uFFFA%v\uFFFB"*/ "<ruby>%v<rt>%v</rt></ruby>"
			if yomi==writing && yomi=="々" {fmt.Fprintf(bw, format, yomi, " "); continue}
			for parts, j := diff.Do(yomi,writing), 0; j<len(parts); {
				part := parts[j]
				switch {
				case part.Added:
					j++
					var yomi diff.Component
					if j<len(parts) && parts[j].Removed {
						yomi = parts[j]
						j++
					}
					fmt.Fprintf(bw, format,
						part, func() string {
							if len(yomi.Value)==1 && yomi.Value[0]=='*' {return " "}; return yomi.String()
						} ())
				case part.Removed:
					panic(fmt.Sprintf("unexpected \"removed\" component: %v in %#v\n", part, line))
				default:
					fmt.Fprint(bw, part)
					j++
				}
			}
		}
		fmt.Fprintln(bw, line[i:]); handleLineFlush()
	}
}
