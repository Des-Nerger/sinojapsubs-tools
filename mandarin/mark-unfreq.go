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
	defer os.Exit(0)
	frequentWordsSet := map[string]struct{}{}
	{
		stdinSc := bufio.NewScanner(os.Stdin)
		for stdinSc.Scan() {
			frequentWordsSet[ strings.Split(stdinSc.Text(),"\t")[1] ] = struct{}{}
		}
	}
	var updateDictionary bool
	flag.BoolVar(&updateDictionary, "update-dictionary", false, "")
	var outputFileName string
	flag.StringVar(&outputFileName, "o", "", "")
	flag.Parse()
	inputFileName := flag.Arg(0)
	dictionary, isSkippedRepetitions := new(Dictionary).Init(flag.Arg(1))
	addedWordsCount := 0
	subs, err := astisub.OpenFile(inputFileName); FatalCheck(err)
	func() {
		nlpir, err := gonlpir.NewNLPIR(gonlpir.UTF8, ""); FatalCheck(err)
		defer nlpir.Exit()
		nlpir.ImportUserDict(flag.Arg(2), true)
		keywords := regexp.MustCompile(flag.Arg(3))
		for _, item := range subs.Items {
			itemString := item.String()
			keywordBoundaries := []int(nil)
			{
				locs := keywords.FindAllStringIndex(itemString, -1)
				keywordBoundaries = make([]int, len(locs)*2+1)
				for i, loc := range locs {
					keywordBoundaries[2*i], keywordBoundaries[2*i+1] = loc[0], loc[1]
				}
				keywordBoundaries[len(keywordBoundaries)-1] = -1
			}
			results := nlpir.ParagraphProcessA(itemString, true)
			sb := strings.Builder{}
			processedBytesCount := 0
			unhideKeywordsWrite := func(word string) {
				for i:=0; i<len(word); i++ {
					switch word[i] {
					case '<', '>': // Do nothing
					default:
						for processedBytesCount == keywordBoundaries[0] {
							sb.WriteByte('|')
							keywordBoundaries = keywordBoundaries[1:]
						}
						processedBytesCount++
					}
					sb.WriteByte(word[i])
				}
			}
			for _, result := range results {
				//sb.WriteRune(' ')
				if result.Spos != "" {
					switch result.Spos[0] {
					case 'w', 'm': // Do nothing
					default:
						if _, ok := frequentWordsSet[result.Word]; !ok {
							if translation, ok := dictionary.Translate(result); ok {
								if translation != "" {
									unhideKeywordsWrite(translation)
									continue
								}
							} else {
								dictionary.AddWord(result.Word, "")
								addedWordsCount++
							}
							sb.WriteByte('<')
							unhideKeywordsWrite(result.Word)
							sb.WriteByte('>')
							continue
						}
					}
				}
				unhideKeywordsWrite(result.Word)
			}
			for processedBytesCount == keywordBoundaries[0] {
				sb.WriteByte('|')
				keywordBoundaries = keywordBoundaries[1:]
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

	if updateDictionary {
		defer fmt.Fprintf(os.Stdout, "addedWordsCount = %v\n", addedWordsCount)
		if addedWordsCount>0 || isSkippedRepetitions {
			FatalCheck(dictionary.WriteToFile(flag.Arg(1)))
		}
	}
}
