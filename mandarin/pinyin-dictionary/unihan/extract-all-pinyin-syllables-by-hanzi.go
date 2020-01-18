package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()

	hanziToSyllables := map[rune][]string{}
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "\t")
		if len(s)>3 {panic(fmt.Sprintf(">3 fields: %#v", s))}
		if len(s)<3 {continue}
		switch s[1] {case "kHanyuPinyin","kXHC1983","kHanyuPinlu","kMandarin":; default:continue}
		code := s[0]
		if !strings.HasPrefix(code, "U+") {panic("unexpected prefix: " + code)}
		i, err := strconv.ParseInt(code[2:], 16, 32)
		if err!=nil {panic(err)}
		r := rune(i)
		if !unicode.In(r, unicode.Han) {
			panic(fmt.Sprintf("not a Han character: %v", code))
		}
		syllables := hanziToSyllables[r]
		addSyllableIfNotPresent := func(syllable string) {
			if !contains(syllables, syllable) {
				syllables = append(syllables, syllable)
			} else {
				fmt.Fprintf(os.Stderr, "skipped %c's duplicate \"%s\" in: %v\n", r, syllable, s[2])
			}
		}
		switch s[1] {
		case "kHanyuPinyin", "kXHC1983":
			for _, e := range strings.Split(s[2], " ") {
				for _, syllable := range strings.Split(e[strings.IndexByte(e, ':')+1:], ",") {
					addSyllableIfNotPresent(syllable)
				}
			}
		case "kHanyuPinlu":
			for _, e := range strings.Split(s[2], " ") {
				addSyllableIfNotPresent(e[:strings.IndexByte(e, '(')])
			}
		case "kMandarin":
			addSyllableIfNotPresent(strings.Split(s[2], " ")[0])
		}
		hanziToSyllables[r] = syllables
	}

	for r, syllables := range hanziToSyllables {
		bw.WriteRune(r)
		for _, syllable := range syllables {
			bw.WriteByte('\t')
			bw.WriteString(syllable)
		}
		bw.WriteByte('\n')
	}
}

func contains(ss []string, s string) bool {
	for _, ssS := range ss {
		if ssS == s {return true}
	}
	return false
}
