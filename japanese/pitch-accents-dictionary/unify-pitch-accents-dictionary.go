package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/miiton/kanaconv"
)

func main() {
	contains := func(s string, ranges ...*unicode.RangeTable) bool {
		for _, r := range s {
			if unicode.In(r, ranges...) {return true}
		}
		return false
	}
	fixIfNoReading := func(fields []string) (ok bool) {
		if fields[1] == "" {
			if !contains(fields[0], unicode.Han) {
				fields[1] = fields[0]
			} else {
				fmt.Fprintf(os.Stderr, "unfixable \"no reading\" situation at: %#v\n", fields)
				return false
			}
		}
		return true
	}

	type unifiedValue struct {
		definitions []string
		pitchAccents []string
	}
	unifiedDict := map[[2]string]*unifiedValue{}

	func () {
		file, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			e := parseJmdictEntry(scanner.Text())
			if !fixIfNoReading(e) {continue}
			key := [...]string{e[0], /*kanaconv.ZenkakuToHankaku*/(kanaconv.HiraganaToKatakana(e[1]))}
			v, ok := unifiedDict[key]
			if !ok {
				v = &unifiedValue{}
				unifiedDict[key] = v
			}
			v.definitions = append(v.definitions, e[2])
		}
	} ()

	func () {
		file, err := os.Open(os.Args[2])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		re := regexp.MustCompile(`[(][^)]*[)]|[)]|[?]|[|]Ø`)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := strings.Split(scanner.Text(), "\t")
			if !fixIfNoReading(s) {continue}
			key := [...]string{s[0], /*kanaconv.ZenkakuToHankaku*/(kanaconv.HiraganaToKatakana(s[1]))}
			v, ok := unifiedDict[key]
			if !ok {
				v = &unifiedValue{}
				unifiedDict[key] = v
			}
			v.pitchAccents = append(v.pitchAccents, re.ReplaceAllLiteralString(s[2], ""))
		}
	} ()

	removeAdjacentDuplicates := func(ss []string) []string {
		i := 0
		for j:=i+1; j<len(ss); j++ {
			if ss[i] != ss[j] {
				i++
				ss[i] = ss[j]
			}
		}
		return ss[:i+1]
	}
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	for k, v := range unifiedDict {
		bw.WriteString(k[0])
		bw.WriteByte('\t')
		bw.WriteString(k[1])
		bw.WriteByte('\t')
		{
			p := v.pitchAccents
			if len(p) > 1 {
				fmt.Fprintf(os.Stderr, "%v --> ", p)
				p = removeAdjacentDuplicates(p)
				if len(p)>1 {os.Stderr.WriteString("!!!!!!!!!!!!!!!!!!!!!!!!!!!!! ")}
				fmt.Fprintf(os.Stderr, "%v\n", p)
			}
			bw.WriteString(strings.Join(p, "|"))
		}
		bw.WriteByte('\t')
		bw.WriteString(strings.Join(v.definitions, "/"))
		bw.WriteByte('\n')
	}
}

func parseJmdictEntry(s string) [/*3*/]string {

	p, i, fieldStart := make([]string, 3), 0, 0
	unexpectedPrefixPanic := func() {panic(fmt.Errorf("unexpected prefix after %v in \"%v\"", i, s))}
loop:
	for j:=0; j<len(s); j++ {
		b := s[j]
		switch i {
		case 0:
			switch b {
			case '[':
				const sep = " ["
				sepStart := (j+1)-len(sep)
				if !strings.HasPrefix(s[sepStart:], sep) {
					unexpectedPrefixPanic()
				}
				p[i] = s[fieldStart:sepStart]
				fieldStart = sepStart + len(sep)
				i = 1
			case '/':
				const sep = " /"
				sepStart := (j+1)-len(sep)
				if !strings.HasPrefix(s[sepStart:], sep) {
					unexpectedPrefixPanic()
				}
				p[i] = s[fieldStart:sepStart]
				fieldStart = sepStart + len(sep)
				i = 2
			}
		case 1:
			switch b {
			case ']':
				const sep = "] /"
				if !strings.HasPrefix(s[j:], sep) {
					unexpectedPrefixPanic()
				}
				p[i] = s[fieldStart:j]
				fieldStart = j+len(sep)
				i = 2
			}
		case 2:
			break loop
		}
	}
	if fieldStart < len(s) {
		if s[len(s)-1] != '/' {panic(fmt.Errorf("unexpected terminator after %v in \"%v\"", i, s))}
		p[i] = s[fieldStart:len(s)-1]
	}
	return p
}
