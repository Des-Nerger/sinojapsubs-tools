package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	. "github.com/Des-Nerger/sinojapsubs-tools/commonrangetables"
)

type Dictionary map[string]string

func (d *Dictionary) Init(fileName string) *Dictionary {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	*d = make(Dictionary)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {continue}
		splits := strings.SplitN(line, "\t", 2)
		spaceIfNeeded := func(r rune, _ int) string {
			if unicode.Is(AsciiAlphaNum, r) {
				return " "
			}
			return ""
		}
		(*d)[splits[0]] =
			spaceIfNeeded(utf8.DecodeRuneInString(splits[1])) +
			splits[1] +
			spaceIfNeeded(utf8.DecodeLastRuneInString(splits[1]))
	}
	return d
}

func (d *Dictionary) Translate(morphemes ...Morpheme) (translation string, ok bool) {
	morpheme := morphemes[0]
	if translation, ok = (*d)[morpheme[1]]; !ok {
		const (
			asciiAlphaNum = iota != 0
			alphaNum
		)
		state := asciiAlphaNum
		for _, r := range morpheme[1] {
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
			return " " + morpheme[1] + " ", true
		}
		return morpheme[1], true
	}
	translation += func() string {
		switch {
		case strings.HasPrefix(morpheme[3], "ナ形容詞"),
		     morpheme[3]=="ナノ形容詞":
			return strings.TrimPrefix(morpheme[0], morpheme[1])
		case strings.HasPrefix(morpheme[3], "イ形容詞"):
			return strings.TrimPrefix(morpheme[0], strings.TrimSuffix(morpheme[1], "い"))
			//fmt.Fprintf(os.Stdout, "[%q, %q] translation: %q\n", morpheme[0], morpheme[1], translation)
		case strings.Contains(morpheme[3], "動詞"):
			if len(morphemes) > 1 {
				switch morphemes[1][1] {
				case "する", "できる":
					return ""
				}
			}
			switch morpheme[4] {
			case "タ形": // past
				return "した"
			case "タ系連用タリ形": // -tari form
				return "したり"
			case "タ系連用チャ形": // -chau form
				return "しちゃ"
			case "タ系連用テ形": // -te form
				return "して"
			case "基本形": // dictionary form
				return "する"
			case "未然形": // imperfective
			dummyLoop:
				for {
					if len(morphemes) > 1 {
						nextMorpheme := &morphemes[1]
						switch nextMorpheme[1] {
						case "させる", "られる":
							for i:=0; i<2; i++ {
								_, size := utf8.DecodeRuneInString(nextMorpheme[i])
								nextMorpheme[i] = nextMorpheme[i][size:]
							}
							fallthrough
						case "せる", "れる":
							return "さ"
						case "ぬ":
							return "せ"
						case "ない":
							break dummyLoop
						default:
							switch nextMorpheme[0] {
							case "ん", "の":
								break dummyLoop
							}
						}
					}
					fmt.Fprintf(os.Stdout, "unknown imperfective:\n\t%q\n", morphemes[: min(2,len(morphemes))])
					break
				}
				fallthrough
			case "基本連用形", // masu stem
			     "タ接連用形": // タ接 form
				return "し"
			case "意志形": // volitional
				return "しよう"
			case "命令形": // imperative
				return "しろ"
			case "文語命令形": // literary imperative
				return "せよ"
			case "基本条件形": // provisional/general conditional -eba
				return "すれば"
			case "タ系条件形": // past conditional -tara
				return "したら"
			case "文語連体形": // literary attributive
				if morpheme[3] == "助動詞く型" {
					return "き"
				}
				fallthrough
			default:
				fmt.Fprintf(os.Stdout, "unknown verb form: %q\n", morpheme)
				ok=false
			}
		default: // Do nothing
			if morpheme[2] == "動詞" {
				fmt.Fprintf(os.Stdout, "morheme[3] of a verb doesn't contain \"動詞\": %q\n", morpheme)
			}
			if morpheme[0] != morpheme[1] {
				fmt.Fprintf(os.Stdout, "%v: %q != %q\n", morpheme[2], morpheme[0], morpheme[1])
				//ok=false
			}
		}
		return ""
	} ()
	return
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
