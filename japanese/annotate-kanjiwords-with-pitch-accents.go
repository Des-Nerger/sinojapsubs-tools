package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	// "runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
	"github.com/miiton/kanaconv"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func fatalCheck(err error) {
	if err != nil {
		panic(err)
	/*
		fmt.Fprintln(os.Stderr, err)
		//os.Exit(1)
		runtime.Goexit()
	*/
	}
}

type pitchAccentedReading [2]string

func main() {
	//defer os.Exit(0)

	var outputFilename string
	flag.StringVar(&outputFilename, "o", "", "")
	flag.Parse()

	inputFilename := flag.Arg(0)
	subs, err := astisub.OpenFile(inputFilename); fatalCheck(err)

	pitchAccentedReadings := map[string][]pitchAccentedReading{}
	func () {
		file, err := os.Open(flag.Arg(1))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			s := strings.Split(line, "\t")
			if s[2]=="" {continue}
			pitchAccentedReadings[s[0]] = append(pitchAccentedReadings[s[0]], pitchAccentedReading{s[1], s[2]})
		}
	} ()
/*
	for writing, readings := range pitchAccentedReadings {
		if len(readings) >= 6 {
			fmt.Printf("%v\t%v", len(readings), writing)
			for _, reading := range readings {
				fmt.Printf("\t%v%v", reading[0], reading[1])
			}
			fmt.Println()
		}
	}
	return
*/

	kanjiwordsAnnotationExceptions := map[string]struct{}{}
	func () {
		file, err := os.Open(flag.Arg(2))
		if err != nil {
			panic(err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line=="" || line[0]=='#' {continue}
			if line!=strings.TrimSpace(line) {
				panic(fmt.Errorf("\"%v\": trimmable whitespace: \"%v\"", flag.Arg(2), line))
			}
			kanjiwordsAnnotationExceptions[line] = struct{}{}
		}
	} ()

	const (
		irrelevant = iota
		katakanaOrAsciiDigit
		prefixHiragana
		nonprefixHiragana
	)
	var (
		previousLastRuneKind int
		morpheme Morpheme
		sb = strings.Builder{}
	)
	writeDelimitString := func(s string, hide bool) {
		firstRune, _ := utf8.DecodeRuneInString(s)
		lastRune, _ := utf8.DecodeLastRuneInString(s)
		isPrefix := morpheme[2]=="接頭辞"
		for i, r := range [...]rune{firstRune, lastRune} {
			runeKind := func() int {
				switch {
				case '0' <= r && r <= '9', unicode.In(r, unicode.Katakana):
					return katakanaOrAsciiDigit
				case unicode.In(r, unicode.Hiragana):
					if isPrefix {return prefixHiragana}
					return nonprefixHiragana
				}
				return irrelevant
			} ()
			if i==1 {
				previousLastRuneKind = runeKind
				continue
			}
			switch previousLastRuneKind {
			case nonprefixHiragana:
				switch runeKind {
				case katakanaOrAsciiDigit, prefixHiragana:
					sb.WriteByte(' ')
				}
			case katakanaOrAsciiDigit:
				switch runeKind {
				case katakanaOrAsciiDigit:
					sb.WriteRune('・')
				case prefixHiragana:
					sb.WriteRune(' ')
				}
			}
			for j:=false; ; j=!j {
				if hide {sb.WriteByte('|')}
				if j {break}
				sb.WriteString(s)
			}
		}
	}

	jumanpp := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
	jumanpp.Start()
	defer jumanpp.Wait()
	for _, item := range subs.Items {
		for i:=0; i<len(item.Lines); i++ {
			line := &item.Lines[i]
			if len(line.Items) > 1 {
				fmt.Fprintf(os.Stderr, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
				for _, lineItem := range line.Items {
					fmt.Fprintf(os.Stderr, "%q\n", lineItem.Text)
				}
				fmt.Fprintln(os.Stderr)
				continue
			}
			sb.Reset()
			previousLastRuneKind = irrelevant
			li := &line.Items[0]
			annotations := []string{"#"}
			for _, morpheme = range jumanpp.AnalyzeLine(line.Items[0].Text) {
				switch morpheme[2] {
				default:
					if !contains(morpheme[1], unicode.Han) {break}
					if _, ok := kanjiwordsAnnotationExceptions[morpheme[1]]; ok {break}
					ps, ok := pitchAccentedReadings[morpheme[1]]
					if !ok {break}
					conjugated := conjugatePitchAccentedReadingsAsIn(morpheme, ps)
					writeDelimitString(func() (string, bool) {
						if conjugated==morpheme[0] || len(ps)==1 {
							return conjugated, false
						}
						annotations = append(annotations, conjugated)
						return morpheme[0], true
					} ())
					continue
				case "特殊", /*"未定義語",*/ "0":
				}
				writeDelimitString(morpheme[0], false)
			}
			li.Text = sb.String()
			if len(annotations)>=2 {
				item.Lines = insert(item.Lines, i+1,
					astisub.Line{Items: []astisub.LineItem{{Text: strings.Join(annotations, "   ")}}},
				)
				i++
			}
		}
	}

	if outputFilename == "" {
		outputFilename = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFilename, "$2.anoted.srt")
		if inputFilename == outputFilename {
			fatalCheck(errors.New("inputFilename == outputFilename"))
		}
	}
	err = subs.Write(outputFilename); fatalCheck(err)
}

func contains(s string, ranges ...*unicode.RangeTable) bool {
	for _, r := range s {
		if unicode.In(r, ranges...) {return true}
	}
	return false
}

func insert(lines []astisub.Line, i int, line astisub.Line) []astisub.Line {
	lines = append(lines[:i], lines[i-1:]...)
	lines[i] = line
	return lines
}

func conjugatePitchAccentedReadingsAsIn(m Morpheme, ps []pitchAccentedReading) string {
	const kanaRuneLen = 3
	dictFormSuffixLen := func () int {
		switch {
		case strings.HasPrefix(m[3], "イ形容詞"), strings.Contains(m[3], "動詞"):
			return kanaRuneLen
		}
		return 0
	} ()
	conjugatedSuffix := m[0][len(m[1])-dictFormSuffixLen:]
	prefixLen := len(m[5])-len(conjugatedSuffix)
	conjugatedReading := kanaconv.HiraganaToKatakana(m[5][:prefixLen]) + m[5][prefixLen:]

	prefixes := make([]string, len(ps))
	jumanppChosenIndex := -1
	for i, p := range ps {
		kana := p[0]
		prefix := kana[:len(kana)-dictFormSuffixLen]
		if jumanppChosenIndex==-1 && prefix+conjugatedSuffix==conjugatedReading {
			jumanppChosenIndex = i
		} 
		prefixes[i] = prefix + p[1]
	}
	if jumanppChosenIndex==-1 {return m[0]}
	last := len(prefixes)-1
	prefixes[last], prefixes[jumanppChosenIndex] = prefixes[jumanppChosenIndex], prefixes[last]
	return strings.Join(prefixes, "/") + conjugatedSuffix
}
