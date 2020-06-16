package main
import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

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
	const maxInt = 1<<(strconv.IntSize-1)-1
	defer os.Exit(0)
	var outputFileName string
	flag.StringVar(&outputFileName, "o", "", "")
	flag.Parse()
	inputFileName := flag.Arg(0)
	subs, err := astisub.OpenFile(inputFileName); FatalCheck(err)
	func() {
		nlpir, err := gonlpir.NewNLPIR(gonlpir.UTF8, ""); FatalCheck(err)
		defer nlpir.Exit()
		nlpir.ImportUserDict(flag.Arg(1), true)
		keywords := regexp.MustCompile(flag.Arg(2)); keywords.Longest()

		pinyinDictionary := map[string]string{}
		{
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				s := strings.Split(scanner.Text(), "\t")
				if len(s)!=2 {FatalCheck(fmt.Errorf("len(%#v)!=2", s))}
				pinyinDictionary[s[0]] = s[1]
			}
		}

		for _, item := range subs.Items {
			itemString := item.String()
			//FLCVR: locs := append(unifyAdjacent( keywords.FindAllStringIndex(itemString, -1) ), []int{maxInt, maxInt})
			results := nlpir.ParagraphProcessA(itemString, true)
			sb := strings.Builder{}
			processedBytesCount := 0
			previousWasPinyin := false
			for _, result := range results {
				if result.Spos=="" || result.Spos[0]=='w' /*|| result.Spos[0]=='m'*/ {
					sb.WriteString(result.Word)
					previousWasPinyin = false
				} else {
					if keywords.MatchString(result.Word) /*FLCVR: fullyCoveredBy(&locs, processedBytesCount, len(result.Word))*/ {
						sb.WriteByte('<')
						sb.WriteString(result.Word)
						sb.WriteByte('>')
						previousWasPinyin = false
					} else if pinyin, ok := pinyinDictionary[result.Word]; ok {
						if previousWasPinyin {sb.WriteByte(' ')}
						sb.WriteString(pinyin)
						previousWasPinyin = true
					} else {
						//sb.WriteString("<"/*"<|"*/)
						sb.WriteString(result.Word)
						//sb.WriteString(">"/*"|>"*/)
						previousWasPinyin = false
					}
				}
				processedBytesCount += len(result.Word)
			}
			lineStrings := strings.Split(sb.String(), "\n")
			item.Lines = make([]astisub.Line, len(lineStrings))
			for i := range item.Lines {
				item.Lines[i].Items = []astisub.LineItem{{Text: lineStrings[i]}}
			}
		}
	} ()
	if outputFileName == "" {
		outputFileName = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFileName, "$2.marked.srt")
		if inputFileName == outputFileName {
			FatalCheck(errors.New("inputFileName == outputFileName"))
		}
	}
	FatalCheck(subs.Write(outputFileName))
}

/*FLCVR:
func unifyAdjacent(locs [][]int) [][]int {
	for i:=1; i<len(locs); {
		if locs[i-1][1] == locs[i][0] {
			locs[i-1][1] = locs[i][1]
			locs = append(locs[:i], locs[i+1:]...)
			continue
		}
		i++
	}
	return locs
}

func fullyCoveredBy(locs *[][]int, start, length int) bool {
	for start >= (*locs)[0][1] {
		*locs = (*locs)[1:]
	}
	return (*locs)[0][0] <= start && start+length <= (*locs)[0][1]
}
//*/
