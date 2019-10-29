package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Des-Nerger/gonlpir"
	. "github.com/Des-Nerger/sinojapsubs-tools/commonrangetables"
)

type Dictionary struct {
	Map map[string]string
	Slice []string
}

func isSkippedLine(line string) bool {
	return line == "" || line[0] == '#'
}

func (d *Dictionary) Init(fileName string) (_ *Dictionary, isSkippedRepetitions bool) {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d.Map = map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if isSkippedLine(line) {
			d.Slice = append(d.Slice, line)
			continue
		}
		splits := strings.SplitN(line, "\t", 2)
		spaceIfNeeded := func(r rune, _ int) string {
			if unicode.Is(AsciiAlphaNum, r) {
				return " "
			}
			return ""
		}
		if _, ok := d.Map[splits[0]]; ok {
			fmt.Fprintf(os.Stdout, "warning: skipped repetition: [%v]%q\n", splits[0], splits[1])
			isSkippedRepetitions = true
			continue
		}
		d.Map[splits[0]] =
			spaceIfNeeded(utf8.DecodeRuneInString(splits[1])) +
			splits[1] +
			spaceIfNeeded(utf8.DecodeLastRuneInString(splits[1]))
		d.Slice = append(d.Slice, splits[0])
	}
	return d, isSkippedRepetitions
}

func (d *Dictionary) Translate(results ...*gonlpir.Result) (translation string, ok bool) {
	result := results[0]
	if translation, ok = d.Map[result.Word]; !ok {
		const (
			asciiAlphaNum = iota != 0
			alphaNum
		)
		state := asciiAlphaNum
		for _, r := range result.Word {
			switch state {
			case asciiAlphaNum:
				if unicode.Is(AsciiAlphaNum, r) {
					break
				}
				state = alphaNum
				fallthrough
			case alphaNum:
				if unicode.Is(AlphaNum, r) {
					break
				}
				return
			}
		}
		if state == asciiAlphaNum {
			return " " + result.Word + " ", true
		}
		return result.Word, true
	}
	return
}

func (d *Dictionary) AddWord(word, translation string) {
	if !isSkippedLine(word) {
		d.Map[word] = translation
	}
	d.Slice = append(d.Slice, word)
}

func (d *Dictionary) WriteToFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	bw := bufio.NewWriter(file)
	defer bw.Flush()
	for _, str := range d.Slice {
		fmt.Fprintln(bw, func() string {
			if isSkippedLine(str) {
				return str
			}
			return str + "\t" + strings.TrimSpace(d.Map[str])
		} ())
	}
	return nil
}

/*
func main() {
	dictionary := new(Dictionary).Init(os.Args[1])
	fmt.Println(dictionary.Translate(&Morpheme{"", "人"}))

	for key, value := range *dictionary {
		fmt.Fprintf(os.Stdout, "%q: %q\n", key, value)
	}
}
*/
