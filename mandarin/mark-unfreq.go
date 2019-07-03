package main
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/Des-Nerger/gonlpir"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astisub"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func FatalCheck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		//os.Exit(1)
		runtime.Goexit()
	}
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
	dictionary, isSkippedRepetitions := new(Dictionary).Init(os.Args[2])
	addedWordsCount := 0
	subs, err := astisub.OpenFile(inputFileName); FatalCheck(err)
	func() {
		nlpir, err := gonlpir.NewNLPIR(gonlpir.UTF8, ""); FatalCheck(err)
		defer nlpir.Exit()
		nlpir.ImportUserDict(os.Args[3], true)
		for _, item := range subs.Items {
			itemString := item.String()
			sb := strings.Builder{}
			results := nlpir.ParagraphProcessA(itemString, true)
			for _, result := range results {
				//sb.WriteRune(' ')
				if result.Spos != "" {
					switch result.Spos[0] {
					case 'w', 'm': // Do nothing
					default:
						if _, ok := frequentWordsSet[result.Word]; !ok {
							if translation, ok := dictionary.Translate(result); ok {
								if translation != "" {
									sb.WriteString(translation)
									continue
								}
							} else {
								dictionary.AddWord(result.Word, "")
								addedWordsCount++
							}
							sb.WriteString("<" + result.Word + ">")
							continue
						}
					}
				}
				sb.WriteString(result.Word)
			}
			lineStrings := strings.Split(sb.String(), "\n")
			item.Lines = make([]astisub.Line, len(lineStrings))
			for i := range item.Lines {
				item.Lines[i].Items = []astisub.LineItem{{Text: lineStrings[i]}}
			}
		}
	} ()
	outputFileName := regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
		ReplaceAllString(inputFileName, "$2.marked.srt")
	if inputFileName == outputFileName {
		FatalCheck(errors.New("inputFileName == outputFileName"))
	}
	FatalCheck(subs.Write(outputFileName))

	defer fmt.Fprintf(os.Stdout, "addedWordsCount = %v\n", addedWordsCount)
	if addedWordsCount>0 || isSkippedRepetitions {
		FatalCheck(dictionary.WriteToFile(os.Args[2]))
	}
}
