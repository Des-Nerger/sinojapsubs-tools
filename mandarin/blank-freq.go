package main
import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/Des-Nerger/go-astisub"
	"github.com/Des-Nerger/gonlpir"
	"github.com/asticode/go-astilog"
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
	const blankSequence = "＿" //囗口◻⬜▢◯
	defer os.Exit(0)
	frequentWordsSet := map[string]struct{}{}
	{
		stdinSc := bufio.NewScanner(os.Stdin)
		for stdinSc.Scan() {
			frequentWordsSet[ strings.Split(stdinSc.Text(),"\t")[1] ] = struct{}{}
		}
	}
	inputFileName := os.Args[1] //will panic if len(os.Args)<2
	subs, err := astisub.OpenFile(inputFileName); FatalCheck(err)
	func() {
		nlpir, err := gonlpir.NewNLPIR(gonlpir.UTF8, ""); FatalCheck(err)
		defer nlpir.Exit()
		switch len(os.Args) {
		case 2: // Do nothing
		default:
			fmt.Fprintln(os.Stdout, "error: len(os.Args)>3")
			runtime.Goexit()
		/*
			fmt.Fprintln(os.Stdout, "warning: only the first two arguments are recognized, the rest are ignored")
			fallthrough
		*/
		case 3:
			fmt.Fprintf(os.Stdout, "importing dictionary from %q\n", os.Args[2])
			nlpir.ImportUserDict(os.Args[2], true)
		}
		for _, item := range subs.Items {
			itemString := item.String()
			sb := strings.Builder{}
			results := nlpir.ParagraphProcessA(itemString, true)
			for _, result := range results {
				if _, ok := frequentWordsSet[result.Word]; ok {
					sb.WriteString(strings.Repeat(blankSequence, utf8.RuneCountInString(result.Word)))
					continue
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
	outputFileName := regexp.MustCompile(`^(.*/|)([^/]+)[.]([^.]+)[.][^.]+$`).
		ReplaceAllString(inputFileName, "$2.b$3.srt")
	if inputFileName == outputFileName {
		FatalCheck(errors.New("inputFileName == outputFileName"))
	}
	FatalCheck(subs.Write(outputFileName))
}
