package main
import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
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
	var lineMaxLen int
	const maxInt = 1<<(strconv.IntSize-1) - 1
	handleLineFlush := func() func() {
		var lineBuffered bool
		flag.BoolVar(&lineBuffered, "lineBuffered", false, "")
		flag.IntVar(&lineMaxLen, "lineMaxLen", maxInt, "")
		flag.Parse()
		return func() {if lineBuffered {bw.Flush()}}
	} ()
	blanked := make(map[string]struct{}, flag.NArg())
	for _, a := range flag.Args() {blanked[a] = struct{}{}}
	contains := func(s string, rt *unicode.RangeTable) bool {
		for _, r := range s {if unicode.Is(rt, r) {return true}}; return false
	}
	blanks, max := "", func(a, b int) int {if a>b {return a}; return b}
	//abs := func(i int) int {mask:=i>>(strconv.IntSize-1); return(i + mask)^mask}
	var (i int; splits []int)
	handleLineSplits := func() {j:=len(splits)-1; if i==splits[j] {bw.WriteByte('\n'); splits=splits[:j+2]}}
	for scanner.Scan() {
		line := scanner.Text()
		if !contains(line, unicode.Han) {fmt.Fprintln(bw, line); handleLineFlush(); continue}
		splits = append(splits[:0], 0)
		result, err := tagger.Parse(line); panicCheck(err)
		type s struct{fields []string; int}; ses := []s(nil)
		{
			fields := strings.FieldsFunc(result, func(r rune)bool{return r=='\n'})
			ses = make([]s, 0, len(fields))
			i, currentLineLimit := 0, lineMaxLen
			for _, field := range fields {
				const sep = ",\t"
				if strings.HasPrefix(field, sep[:1]) {continue}
				fields := strings.FieldsFunc(field, func(r rune)bool{return strings.ContainsRune(sep, r)})
				i += strings.Index(line[i:], fields[0]) + len(fields[0])
				ses = append(ses, s{fields,i})
				for {
					if last:=len(splits)-1; i<=currentLineLimit {
						lastRune, _ := utf8.DecodeLastRuneInString(line[:i])
						nextRune, _ := utf8.DecodeRuneInString(line[i:])
						if unicode.Is(unicode.Hiragana, lastRune) && bool(
							nextRune=='お' || nextRune=='ご' || !unicode.Is(unicode.Hiragana,nextRune),
						) {
							//fmt.Fprintf(os.Stderr, "%v: splits[%v] was %v, now %v\n", fields[0], last, splits[last], i)
							splits[last] = i
						}
						break
					} else {
						if splits[last]==0 {splits[last] = ses[len(ses)-2].int}
						currentLineLimit = splits[last] + lineMaxLen
						splits = append(splits, 0)
						//fmt.Fprintf(os.Stderr, "i==%v; currentLineLimit==%v; splits==%v\n", i, currentLineLimit, splits)
					}
				}
			}
		}
		splits[func() int {
			l:=len(splits); start:=func()int{if l>=3{return splits[l-3]};return 0}()
			if l>=2 && len(line)-start<=lineMaxLen {return l-2}; return l-1
		} ()] = maxInt
		//fmt.Fprintln(os.Stderr, splits, cap(splits), len(line))
		splits = splits[:1]
		i = 0
		for _, s := range ses {
			writing, yomi := s.fields[0], kanaconv.KatakanaToHiragana(s.fields[len(s.fields)-2])
			bw.WriteString(line[i:s.int-len(writing)])
			i = s.int
			switch _, ok := blanked[s.fields[len(s.fields)-3]]; {
			case ok:
				runeCount := utf8.RuneCountInString(writing)
				const blank = "　"/*"－"*/
				writtenBlanksLen := len(blank)*runeCount
				if len(blanks)<writtenBlanksLen {blanks=strings.Repeat(blank, max(32, 2*runeCount))}
				writing=blanks[:writtenBlanksLen]
				fallthrough
			case !contains(writing, unicode.Han):
				bw.WriteString(writing)
				handleLineSplits()
				continue
			}
			const format = /*"\uFFF9%v\uFFFA%v\uFFFB"*/ "<ruby>%v<rt>%v</rt></ruby>"
			if yomi==writing && yomi=="々" {fmt.Fprintf(bw, format, yomi, " "); handleLineSplits(); continue}
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
			handleLineSplits()
		}
		fmt.Fprintln(bw, line[i:]); handleLineFlush()
	}
}
