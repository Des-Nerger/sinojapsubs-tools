package main
import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
	//"unicode/utf8"

	//"github.com/miiton/kanaconv"
)

func main() {
	type entry struct {s [3]string; Type bool}
	dict := map[string][]entry{}
	func () {
		file, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := strings.SplitN(scanner.Text(), "\t", 4)
			dict[s[1]] = append(dict[s[1]], entry{[...]string{s[2], s[0], s[3]}, false})
			if s[0]!=s[1] {dict[s[0]] = append(dict[s[0]], entry{[...]string{s[2], s[1], s[3]}, true})}
		}
	} ()

	freq := map[string]int{}
	func () {
		file, err := os.Open(os.Args[2])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := strings.SplitN(scanner.Text(), "\t", 2)
			i, err := strconv.Atoi(s[0])
			if err != nil {panic(fmt.Errorf("error converting \"%v\" to int: %v", s[0], err))}
			freq[s[1]] = i
		}
	} ()

/*
	半角GraphAsLower全角AsUpper := unicode.SpecialCase{
		unicode.CaseRange{Lo: 0x00_00_00_00, Hi: '!'-1, Delta: [...]rune{0, 0, 0}},
		unicode.CaseRange{Lo: '!', Hi: '~', Delta: [...]rune{'！' - '!', 0, 0}},
		unicode.CaseRange{Lo: '~'+1, Hi: '！'-1, Delta: [...]rune{0, 0, 0}},
		unicode.CaseRange{Lo: '！', Hi: '～', Delta: [...]rune{0, '!' - '！', 0}},
		unicode.CaseRange{Lo: '～'+1, Hi: 0xFF_FF_FF_FF, Delta: [...]rune{0, 0, 0}},
	}
	z2h := strings.NewReplacer(
		"０", "0", "１", "1", "２", "2", "３", "3", "４", "4",
		"５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
		"｛", "{", "｝", "}",
	)
*/

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
	/*
		i := 0
		for i<len(line) {
			r, size := utf8.DecodeRuneInString(line[i:])
			if !unicode.In(r, unicode.Hiragana, unicode.Katakana) {break}
			i += size
		}
		katakana := kanaconv.HiraganaToKatakana(line[:i])
		pitchAccent := strings.ToLowerSpecial(半角GraphAsLower全角AsUpper, line[i:]) //z2h.Replace(line[i:])
	*/
		kBuilder, restBuilder := strings.Builder{}, strings.Builder{}
		{
			kBuilderBroken := false
			for _, r := range strings.TrimLeftFunc(line, unicode.IsSpace) {
				switch {
				case unicode.IsSpace(r): r,kBuilderBroken=' ',true; fallthrough
				case kBuilderBroken: fallthrough
				default:
					restBuilder.WriteRune(func() rune {
						if '！' <= r && r <= '～' {return r - ('！' - '!')}
						return r
					} ())
				case r=='・', r=='ー', /*r=='ﾞ', r=='ﾟ',*/ unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han):
					kBuilder.WriteRune(r)
				}
			}
		}
		k := /*kanaconv.KatakanaToHiragana(kanaconv.HankakuToZenkaku(*/kBuilder.String()/*))*/
		rest := strings.SplitN(restBuilder.String(), " ", 2)

		entries := dict[k]
		type result struct{freq int; e entry}
		results := make([]result, 0, len(entries))
		for _, e := range entries { //strings.Contains(e.s[2], rest[1])
			if strings.Contains(e.s[0], rest[0]) && (len(rest)<=1 || func() bool {
				for _,s:=range e.s[1:3] {if strings.Contains(s, rest[1]) {return true}}; return false
			} ()) { results = append(results, result{freq[func()string{if e.Type {return k}; return e.s[1]}()], e}) }
		}
		sort.SliceStable(results, func(i, j int) bool {
			return results[j].freq < results[i].freq
		})
		for _, r := range results {
			fmt.Printf(/*"%-9v%-8v%-12v%v\n"*/ "%v　%v　%v　%v\n", r.e.s[0], r.freq, r.e.s[1], r.e.s[2])
		}
	}
}
