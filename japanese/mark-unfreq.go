package main
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

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
	frequentWordsSet := map[string]struct{}{}
	{
		stdinSc := bufio.NewScanner(os.Stdin)
		for stdinSc.Scan() {
			frequentWordsSet[ strings.Split(stdinSc.Text(),"\t")[1] ] = struct{}{}
		}
	}
	inputFileName := os.Args[1]
	dictionary := new(Dictionary).Init(os.Args[2])
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
									if morpheme[0] != morpheme[1] {
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
	outputFileName := regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
		ReplaceAllString(inputFileName, "$2.marked.srt")
	if inputFileName == outputFileName {
		fatalCheck(errors.New("inputFileName == outputFileName"))
	}
	err = subs.Write(outputFileName); fatalCheck(err)
}
