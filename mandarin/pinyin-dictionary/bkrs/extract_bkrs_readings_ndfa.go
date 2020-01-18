package main
import (
	"bufio"
	"fmt"
	"os"
	//"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type automaton struct {
	stopped bool
	s string
	wordSyllables []string
}
func (on *automaton) Step(hanzi rune) (forks []*automaton) {
	syllables := P[hanzi]
	if len(syllables)==0 {
		on.stopped = true
		return
	}
	for _, syllable := range syllables {
		if strings.HasPrefix(on.s, syllable) {
			forks = append(forks,
				&automaton{
					s: on.s[len(syllable):],
					wordSyllables: append(func() []string {
						if len(forks)==0 {return on.wordSyllables}
						return append(make([]string, 0, cap(on.wordSyllables)+1), on.wordSyllables...)
					} (), syllable),
				},
			)
		}
	}
	if len(forks)==0 {on.stopped=true; return}
	*on = *(forks[0])
	return forks[1:]
}

type Pinyin map[rune][]string
func (p Pinyin) Init(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "\t")
		r, size := utf8.DecodeRuneInString(s[0])
		if size!=len(s[0]) {panic(fmt.Sprintf("RuneCountInString(\"%v\")>1", s[0]))}
		p[r] = append(p[r], s[1:]...)
	}
}
func (p Pinyin) ParseWordByHanzis(hanzis []rune, s string) (wordSyllables []string) {
	automata := []*automaton{&automaton{stopped:false, s:s, wordSyllables:[]string(nil)}}
	for _, hanzi := range hanzis {
		allStopped := true
		for _, automaton := range automata {
			if automaton.stopped {continue}
			forks := automaton.Step(hanzi)
			if !automaton.stopped || len(forks)>0 {allStopped=false}
			automata = append(automata, forks...)
		}
		if allStopped {break}
	}
	minLeftoverLen := len(s)
	for _, automaton := range automata {
		if len(automaton.s)<minLeftoverLen {
			minLeftoverLen = len(automaton.s)
		}
	}
	for _, automaton := range automata {
		if minLeftoverLen==len(automaton.s) {
			if len(wordSyllables)>0 {
				fmt.Fprintf(os.Stderr, "ambiguity: %q, so skipping\n", automata)
				return nil
			}
			wordSyllables = automaton.wordSyllables
		}
	}
/*
	if strings.Contains(s, "[") && len(wordSyllables)==0 {
		fmt.Fprintf(os.Stderr, "a: %q\n", automata)
	}
*/
	return
}
var P = Pinyin{}


func parseReadingsByHanzis(hanzis []rune, s string) (_ [][]string, ok bool) {
	readings := append([][]string(nil), nil)
	const (
		syllable = iota
		skipReading
		separators
		tags
	)
	state := syllable
	i := 0
	afterTag := false
	openTagsCount := 0
	afterSpace := false
	for s!="" {
		last := len(readings)-1
		switch state {
		case syllable:
			if afterTag {
				afterTag = false
				if len(readings[last])>0 {
					readings = append(readings, nil)
					last++
					i=0
				}
			}
			wordSyllables := func() []string {
				if i>=len(hanzis) {
					if afterSpace {
						readings = append(readings, nil)
						last++
						i=0
					} else {
						fmt.Fprintf(os.Stderr, "note: \"%v\"[%v]\t\"%v\"\t%v, skipping\n", string(hanzis), i, s, readings)
						return nil
					}
				}
				afterSpace = false
				return P.ParseWordByHanzis(hanzis[i:], s)
			} ()
			if len(wordSyllables)==0 {
				state = skipReading
				readings = readings[:last]
				break
			}
			i += len(wordSyllables)
			readings[last] = append(readings[last], wordSyllables...)
			for _, wordSyllable := range wordSyllables {
				s = s[len(wordSyllable):]
			}
			state = separators
		case skipReading:
			r, size := utf8.DecodeRuneInString(s)
			switch r {
			case ',', ';', '，':
				state = separators
			default:
				s = s[size:]
			case '[':
				state = tags
			}
		case separators:
			r, size := utf8.DecodeRuneInString(s)
			switch r {
			case ',', ';', '，':
				readings = append(readings, nil)
				i = 0
				fallthrough
			case ' ':
				afterSpace = true
				fallthrough
			case '’', '-', '*':
				s = s[size:]
			case ')':
				fmt.Fprintf(os.Stderr, "note: closParen: %#v\n", readings)
				s = s[size:]
			case '[':
				state = tags
			case '(':
				readings = append(readings,
					append(make([]string, 0, cap(readings[last])),
						readings[last][:len(readings[last])-1]...,
					),
				)
				i--
				s = s[size:]
				fallthrough
			default:
				state = syllable
			}
		case tags:
			switch s[0] {
			case '[':
				const prefix = "[/"
				if strings.HasPrefix(s, prefix) {
					openTagsCount--
					s = s[len(prefix):]
				} else {
					openTagsCount++
					s = s[1:]
				}
			case ']':
				if openTagsCount == 0 {
					state = separators
					afterTag = true
				}
				fallthrough
			default:
				s = s[1:]
			}
		}
	}
	filteredReadings := make([][]string, 0, len(readings))
	for _, reading := range readings {
		if len(reading) != len(hanzis) {
			fmt.Fprintf(os.Stderr, "len(%#v) != len(%q), skipping\n", reading, hanzis)
			continue
		}
		if !validateReading(hanzis, reading) {
			panic(fmt.Sprintf("!validateReading(%q, %q)", hanzis, reading))
		}
		filteredReadings = append(filteredReadings, reading)
	}
	return filteredReadings, len(filteredReadings)>0
}

func validateReading(hanzis []rune, reading []string) bool {
	for i, h := range hanzis {
		valid := false
		for _, validReading := range P[h] {
			if reading[i] == validReading {
				valid = true
				break
			}
		}
		if !valid {return false}
	}
	return true
}

func main() {
	P.Init(os.Args[1])
/*
	fmt.Println(
		validateReading([]rune("冷存储"), []string{"lěng","cún","chú"}),
		validateReading([]rune("冷存储"), []string{"lěng","cún","chǔ"}),
	)
	return
//*/
	scanner := bufio.NewScanner(os.Stdin)
/*
	for scanner.Scan() {
		s := strings.SplitN(scanner.Text(), " ", 2)
		//if len(s)<2 {continue}
		hanzis := []rune(s[0])
		fmt.Printf("%+v\n", P.ParseWordByHanzis(hanzis, s[1]))
	}
	return
//*/
	const (
		hanzis = iota
		pinyin
		rest
	)
	state := hanzis
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	currentHanzis := ""
	articlesProcessed := 0
	for scanner.Scan() {
		line := scanner.Text()
		switch state {
		case hanzis:
			if utf8.RuneCountInString(line) > 1 {
				currentHanzis = line
				state = pinyin
			} else {
				state = rest
			}
		case pinyin:
			switch line {
			case " _", " .":
			default:
				if line[:2] == " -" || !containsOnly(currentHanzis, unicode.Han) {break}
				//if strings.ContainsRune(line, '[') {os.Stderr.WriteString(currentHanzis + "\t" + line[1:] + "\n"); articlesProcessed++; break} else{break}
				articlesProcessed++
				//if articlesProcessed & 0x0fff == 0 {fmt.Fprintf(os.Stderr, "articlesProcessed=%v\n", articlesProcessed)}
				//if articlesProcessed>8192 {fmt.Fprintf(os.Stderr, "@%v\n", currentHanzis)}
				readings, ok := parseReadingsByHanzis([]rune(currentHanzis), strings.ToLower(line[1:]))
				if !ok {os.Stderr.WriteString(currentHanzis + "\t" + line[1:] + "\n"); break}
				bw.WriteString(currentHanzis)
				for _, reading := range readings {
					bw.WriteByte('\t')
					for j, syllable := range reading {
						if j>=1 {bw.WriteByte(' ')}
						bw.WriteString(syllable)
					}
				}
				bw.WriteByte('\n')
			}
			state = rest
		case rest:
			if line=="" {
				state = hanzis
			}
		}
	}
	fmt.Fprintf(os.Stderr, "%v\n", articlesProcessed)
}

func containsOnly(s string, ranges ...*unicode.RangeTable) bool {
	for _, r := range s {
		if !unicode.In(r, ranges...) {return false}
	}
	return true
}
