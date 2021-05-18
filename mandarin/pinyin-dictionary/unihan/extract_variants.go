package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func addIfDoesntExist(runes *[]rune, passedRune rune) {
	for _, r := range *runes {if passedRune == r {return}}
	*runes = append(*runes, passedRune)
}

func main() {
	m := map[rune]*[]rune{}

	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	var targetRunes *[]rune; mergeWithTarget := func(runes []rune) {
		for _, r := range runes {
			addIfDoesntExist(targetRunes, r)
			m[r] = targetRunes
		}
	}
	for scanner.Scan() {
		splits := strings.Split(scanner.Text(), "\t")
		if len(splits)>3 {panic(fmt.Sprintf(">3 fields: %#v", splits))}
		if len(splits)<3 || splits[2]=="" {continue}
		switch splits[1] {
		case "kSimplifiedVariant", "kTraditionalVariant", "kZVariant":
		default: continue
		}
		codes := append(splits[:1], strings.Split(splits[2], " ")...)
		runes:=make([]rune,0,len(codes)); targetRunes=nil
		for _, code := range codes {
			code = strings.TrimSuffix(code, "<kHKGlyph")
			if !strings.HasPrefix(code, "U+") {panic("unexpected prefix: " + code)}
			i, err := strconv.ParseInt(code[2:], 16, 32)
			if err!=nil {panic(err)}
			r := rune(i)
			if !unicode.Is(unicode.Han, r) {panic(fmt.Sprintf("not a Han character: %v", code))}
			runes = append(runes, r)
			if runes:=m[r]; runes!=nil {if targetRunes==nil {targetRunes=runes} else {mergeWithTarget(*runes)}}
		}
		if targetRunes==nil {targetRunes=&[]rune{}}
		mergeWithTarget(runes)
	}
	for r := range m {
		runes := m[r]; if *runes==nil {continue}
		for _, r := range *runes {bw.WriteRune(r)}; bw.WriteByte('\n')
		*runes = nil
	}
}
