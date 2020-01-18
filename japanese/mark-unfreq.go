package main
import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func fatalCheck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		//os.Exit(1)
		runtime.Goexit()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	defer os.Exit(0)
	var append見出し語 bool
	flag.BoolVar(&append見出し語,
		"append-見出し語", true,
		`when marking, append 見出し語 (after "|") when it's different from 表層形`,
	)
	var outputFileName string
	flag.StringVar(&outputFileName, "o", "", "")
	flag.Parse()
	frequentWordsSet := map[string]struct{}{}
	{
		stdinSc := bufio.NewScanner(os.Stdin)
		for stdinSc.Scan() {
			line := stdinSc.Text()
			if line == "" || line[0] == '#' {continue}
			fields := strings.Fields(line)
			if !(len(fields)==1 && fields[0]==line) {
				fmt.Fprintf(os.Stdout, "warning: line contains whitespace: %q, but proceeding anyway\n", line)
			}
			frequentWordsSet[line] = struct{}{}
		}
	}
	inputFileName := flag.Arg(0)
	dictionary := new(Dictionary).Init(flag.Arg(1))
	subs, err := astisub.OpenFile(inputFileName); fatalCheck(err)
	func() {
		j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
		j.Start()
		defer j.Wait()
		for _, item := range subs.Items {
			for _, line := range item.Lines {
				if len(line.Items) > 1 {
					fmt.Fprintf(os.Stdout, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
					for _, item := range line.Items {
						fmt.Fprintf(os.Stdout, "%q\n", item.Text)
					}
					fmt.Fprintln(os.Stdout)
					continue
				}
				sb := strings.Builder{}
				li := &line.Items[0]
				morphemes := j.AnalyzeLine(li.Text)
				for i, morpheme := range morphemes {
					if !append見出し語 {morpheme[0] = unhideGrammarHiragana(morpheme)}
					switch morpheme[2] {
					default:
						if _, ok := frequentWordsSet[morpheme[1]]; !ok {
							translation, _ := dictionary.Translate(morphemes[i : min(i+2,len(morphemes))]...)
							if translation != "" {
								sb.WriteString(translation)
								continue
							}
							sb.WriteString("<" + morpheme[0] +
								(func() string {
									if append見出し語 && morpheme[0] != morpheme[1] {
										//fmt.Fprintf(os.Stdout, "%q\n", morpheme)
										return "|" + morpheme[1]
									}
									return ""
								} ()) +
								">",
							)
							continue
						}
						fallthrough
					case "特殊", "未定義語", "0":
						sb.WriteString(morpheme[0])
					}
				}
				li.Text = sb.String()
			}
		}
	} ()
	if outputFileName == "" {
		outputFileName = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFileName, "$2.marked.srt")
		if inputFileName == outputFileName {
			fatalCheck(errors.New("inputFileName == outputFileName"))
		}
	}
	err = subs.Write(outputFileName); fatalCheck(err)
}

func unhideGrammarHiragana(morpheme Morpheme) string {
	unsuffixed := func() string {
		switch {
		case strings.HasPrefix(morpheme[3], "ナ形容詞"),
		     morpheme[3]=="ナノ形容詞":
			switch morpheme[2] {
			case "助動詞", "接尾辞":
				return ""
			}
			return morpheme[1]
		case strings.HasPrefix(morpheme[3], "イ形容詞"):
			switch morpheme[1] {
			case "ない", "いい", "らしい":
				return ""
			}
			return strings.TrimSuffix(morpheme[1], "い")
		case strings.Contains(morpheme[3], "動詞"):
			switch morpheme[2] {
			case "接尾辞", "助動詞":
				return ""
			}
			switch morpheme[1] {
			case "する", "できる", "ある", "いる", "くる", "いく":
				return ""
			}
			_, size := utf8.DecodeLastRuneInString(morpheme[1])
			s := morpheme[1][:len(morpheme[1])-size]
			r, size := utf8.DecodeLastRuneInString(s)
			if len(s)==size || !strings.ContainsRune("えけげせてねべめれ", r) {return s}
			return s[:len(s)-size]
		default:
			switch morpheme[2] {
			case "特殊", "未定義語", "0":
				return morpheme[0]
			case "接尾辞":
				if !allHiragana(morpheme[0]) {
					return morpheme[0]
				}
				fallthrough
			case "助詞", "接続詞", "助動詞", "判定詞":
				return ""
			case "連体詞":
				s := morpheme[0]
				for _, suffix := range 連体詞Suffixes {
					if strings.HasSuffix(s, suffix) {
						return s[:len(s)-len(suffix)]
					}
				}
				return s
			case "名詞":
				s := morpheme[0]
				switch s {
				case "こと", "て", "ん":
					return ""
				}
			/*
				r, size := utf8.DecodeLastRuneInString(s)
				if len(s)==size || !strings.ContainsRune("いきぎしちにびみり", r) {
					return s
				}
				return s[:len(s)-size]
			*/
				for _, suffix := range 名詞Suffixes {
					if strings.HasSuffix(s, suffix) {
						return s[:len(s)-len(suffix)]
					}
				}
				return s
			case "指示詞":
				s := morpheme[0]
				switch s {
				case "そういう":
					return ""
				}
				for _, suffix := range 指示詞Suffixes {
					if strings.HasSuffix(s, suffix) {
						return s[:len(s)-len(suffix)]
					}
				}
				return s
			case "副詞":
				s := morpheme[0]
			/*
				switch s {
				case "なんて":
					return ""
				}
			*/
				for _, suffix := range 副詞Suffixes {
					if strings.HasSuffix(s, suffix) {
						return s[:len(s)-len(suffix)]
					}
				}
				return s
			case "感動詞":
				switch morpheme[0] {
				case "じゃ":
					return ""
				}
			}
			return morpheme[0]
		}
	} ()
	return unsuffixed + func() string {
		suffix := strings.TrimPrefix(morpheme[0], unsuffixed)
		//fmt.Printf("%#v | %#v @@ %+v\n", unsuffixed, suffix, morpheme)
		if suffix=="" {return ""}
		return "|" + suffix + "|"
	} ()
}

var (
	名詞Suffixes = [...]string{"だ"}
	連体詞Suffixes = [...]string{"が", "なる"}
	指示詞Suffixes = [...]string{"に", "の", "な"}
	副詞Suffixes = [...]string{"にも", "に", "らか", "ら", "で", "て"}
)

func allHiragana(s string) bool {
	for _, r := range s {
		if !unicode.In(r, unicode.Hiragana) {
			return false
		}
	}
	return true
}
