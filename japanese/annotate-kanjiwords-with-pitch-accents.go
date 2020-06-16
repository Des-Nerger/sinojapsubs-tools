package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	// "runtime"
	"strconv"
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
const kanaRuneLen = 3

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
		countMorae := func(s string) (count int) {
			for _, r := range s {
				switch r {
				case 'ァ','ィ','ゥ','ェ','ォ',
				     'ャ',    'ュ',    'ョ',
				     'ヮ',
				     '・':
				//case 'ヵ','ヶ': fmt.Fprintf(os.Stderr, "+%v+\n", s); fallthrough
				default:
					count++
				}
			}
			//if count<=3 && count != len(s)/kanaRuneLen {fmt.Fprintf(os.Stderr, "_%v_%v\n", s, count)}
			return
		}
		for scanner.Scan() {
			line := scanner.Text()
			s := strings.Split(line, "\t")
			if s[2]=="" && countMorae(s[1])<3 {continue}
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
		spaceOrNone = iota
		katakana
		digit
		prefixHiragana
		nonprefixHiragana
		otherNonspace
	)
	var (
		previousFirstRuneIsDigit bool
		previousLastRuneKind int
		morpheme Morpheme
		sb = strings.Builder{}
	)
	writeDelimitString := func(s string, hide bool) {
		firstRune, _ := utf8.DecodeRuneInString(s)
		lastRune, _ := utf8.DecodeLastRuneInString(s)
		isPrefix := morpheme[2]=="接頭辞"
		firstRuneIsDigit := false
		for i, r := range [...]rune{firstRune, lastRune} {
			runeKind := func() int {
				switch {
				case unicode.IsSpace(r): return spaceOrNone
				case unicode.In(r, unicode.Digit):
					if i==0 {
						firstRuneIsDigit = true
					}
					return digit
				case unicode.In(r, unicode.Katakana): return katakana
				case unicode.In(r, unicode.Hiragana):
					if isPrefix {return prefixHiragana}
					return nonprefixHiragana
				}
				return otherNonspace
			} ()
			if i==1 {
				previousFirstRuneIsDigit = firstRuneIsDigit
				previousLastRuneKind = runeKind
				continue
			}
			switch previousLastRuneKind {
			case nonprefixHiragana:
				switch runeKind {
				case katakana, digit, prefixHiragana:
					sb.WriteByte(' ')
				}
			case katakana:
				switch runeKind {
				case katakana:
					sb.WriteRune('・')
				case prefixHiragana, digit:
					sb.WriteRune(' ')
				}
			case digit:
				switch runeKind {
				case katakana:
					if !previousFirstRuneIsDigit {
						sb.WriteRune('・')
					}
				case prefixHiragana, digit:
					sb.WriteRune(' ')
				}
			/*
			case prefixHiragana:
				if runeKind==digit {
					panic(fmt.Sprintf("hiragana prefix before digit: «%v» «%v»", sb.String(), s))
				}
			*/
			case otherNonspace:
				if runeKind==digit {
					sb.WriteRune(' ')
				}
			}
		/*
			for j:=false; ; j=!j {
				if hide {sb.WriteByte('|')}
				if j {break}
				sb.WriteString(s)
			}
		*/
			sb.WriteString(func() string {
				if !hide {return s}
				return strings.Map(func(r rune) rune {
					if unicode.Is(unicode.Han, r) {
						return '＿'
					}
					return r
				}, s)
			} ())
		}
	}

	jumanpp := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
	jumanpp.Start()
	defer jumanpp.Wait()
	eToU := map[string]string {"ね": "ぬ",
		"え": "う", "て": "つ", "れ": "る", "け": "く",
		"げ": "ぐ", "べ": "ぶ", "め": "む", "せ": "す",
	}
	//UNEXOC: unexceptedKanjiWordsOccurences := map[string]int{}
	for itemIndex, item := range subs.Items {
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
			previousFirstRuneIsDigit = false
			previousLastRuneKind = spaceOrNone
			li := &line.Items[0]
			annotations := []string{"#"}
			for _, morpheme = range jumanpp.AnalyzeLine(line.Items[0].Text) {
			switchLabel:
				switch morpheme[2] {
				default:
					if !contains(morpheme[1], unicode.Han) {break}
					var (ps []pitchAccentedReading; ok bool) 
					for j:=0;; j++ {
						if _, ok = kanjiwordsAnnotationExceptions[morpheme[1]]; ok {break switchLabel}
						//UNEXOC: unexceptedKanjiWordsOccurences[morpheme[1]]++; break switchLabel
						ps, ok = pitchAccentedReadings[morpheme[1]]
						if ok {
							//if j==1 {fmt.Fprintf(os.Stderr, "used «%v»\n", morpheme[1])}
							break
						} //else {if j==1 {fmt.Fprintf(os.Stderr, "not used «%v»\n", morpheme[1])}}
						if j==1 || !strings.Contains(morpheme[3], "動詞") {break /*switchLabel*/} //
						penultimate := len(morpheme[1]) - 2*kanaRuneLen
						u, ok := eToU[morpheme[1][penultimate:penultimate+kanaRuneLen]]
						if !ok {break /*switchLabel*/} //
						fmt.Fprintf(os.Stderr, "deconjugating %v into ", morpheme[1])
						morpheme[1] = morpheme[1][:penultimate] + u
						fmt.Fprintln(os.Stderr, morpheme[1])
					}
					conjugated := conjugatePitchAccentedReadingsAsIn(morpheme, ps)
					if conjugated == "" {break}
					//if morpheme[1]=="駆ける" {fmt.Fprintf(os.Stderr, "%q\n%q\n", ps, conjugated)}
					writeDelimitString(func() (string, bool) {
						if strings.Count(conjugated, "/")==0 {
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
		item.Lines = append(
			[]astisub.Line{{Items: []astisub.LineItem{{Text: strconv.Itoa(itemIndex+1)}}}},
			item.Lines...,
		)
	}

	//UNEXOC: for k,v := range unexceptedKanjiWordsOccurences {fmt.Fprintf(os.Stdout, "%v\t%v\n", v, k)}
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


var workaroundRegexp = regexp.MustCompile("[ー～]っ?$") //"ーっ" "～っ"
func conjugatePitchAccentedReadingsAsIn(m Morpheme, ps []pitchAccentedReading) string {
	var workaround string
	dictFormSuffixLen := kanaRuneLen * func () int {
		switch {
		case strings.Contains(m[3], "動詞"):
			workaround = workaroundRegexp.FindString(m[0])
			m[0] = m[0][:len(m[0])-len(workaround)]
			switch m[3] {case "ザ変動詞", "サ変動詞": return 2}
			fallthrough
		case strings.HasPrefix(m[3], "イ形容詞"): fallthrough
		case m[2]=="連体詞" && func() bool {
			r, _ := utf8.DecodeLastRuneInString(m[1])
			return unicode.Is(unicode.Hiragana, r)
		} ():
			return 1
		case m[3]=="タル形容詞":
			return 2
		}
		return 0
	} ()
	//fmt.Fprintf(os.Stderr, "%v ||| %v\n", m, dictFormSuffixLen)
	conjugatedSuffix := m[0][len(m[1])-dictFormSuffixLen:]
	prefixLen := len(m[5])-len(conjugatedSuffix) /* may work dirty,
		as in [起ーきーろー 起きる 動詞 母音動詞 命令形 おきろ] ||| 3
	*/
	conjugatedReading := kanaconv.HiraganaToKatakana(m[5][:prefixLen]) + m[5][prefixLen:]

	prefixes := make([]string, len(ps), len(ps)+1)
	jumanppChosenIndex := -1
	for i, p := range ps {
		kana := p[0]
		prefix := kana[:len(kana)-dictFormSuffixLen]
		if jumanppChosenIndex==-1 && prefix+conjugatedSuffix==conjugatedReading {
			jumanppChosenIndex = i
		} 
		prefixes[i] = prefix + p[1]
	}
	if jumanppChosenIndex==-1 {
		//return m[0]
		jumanppChosenPrefix := conjugatedReading[:prefixLen]
		if !contains(jumanppChosenPrefix, unicode.Han) {
			prefixes = append(prefixes, jumanppChosenPrefix)
		}
	} else {
		prefixes = append(prefixes, prefixes[jumanppChosenIndex])
		prefixes = append(prefixes[:jumanppChosenIndex], prefixes[jumanppChosenIndex+1:]...)
	}
	return strings.Join(prefixes, "/") +
		func() string {if len(prefixes)==1 || conjugatedSuffix=="" {return ""}; return "・"} () +
		conjugatedSuffix + workaround
}
