package main
import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func main() {
	freq := map[string]uint{}
	var occurrencesTotal uint64
	func() {
		j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
		j.Start()
		defer j.Wait()
		filenames := []string(nil)
		if len(os.Args) <= 1 {
			sc := bufio.NewScanner(os.Stdin)
			for sc.Scan() {
				filenames = append(filenames, sc.Text())
			}
		} else {
			filenames = os.Args[1:]
		}
		for _, filename := range filenames {
			fmt.Fprintln(os.Stdout, filename)
			subs, err := astisub.OpenFile(filename)
			if err != nil {
				fmt.Fprintln(os.Stdout, err)
				continue
			}
			for _, item := range subs.Items {
				for _, line := range item.Lines {
					lineString := func() string {
						if len(line.Items) > 1 {
							var texts []string
							for _, lineItem := range line.Items {
								texts = append(texts, lineItem.Text)
							}
							return strings.Join(texts, "")
						}
						return line.Items[0].Text
					} ()
/*
					if len(line.Items) > 1 {
						fmt.Fprintf(os.Stdout, "len(line.Items)==%v: %q\n", len(line.Items), lineString)
					}
*/
					for _, morpheme := range j.AnalyzeLine(lineString) {
						switch morpheme[2] {
						case "特殊", "未定義語", "0": // Do nothing
						default:
							freq[morpheme[1]]++
							occurrencesTotal++
						}
					}
				}
			}
		}
	} ()
	type pair struct{uint; string}
	freqPairs := make([]pair, len(freq))
	{
		i := 0
		for word, count := range freq {
			freqPairs[i] = pair{uint:count, string:word}
			i++
		}
	}
	sort.Slice(freqPairs, func(i, j int) bool {return freqPairs[j].uint < freqPairs[i].uint})
	for _, pair := range freqPairs {
		fmt.Fprintf(os.Stderr, "%v\t%v\n", pair.uint, pair.string)
	}
/*
	words := make([]string, len(freq))
	{
		i := 0
		for word := range freq {
			words[i] = word
			i++
		}
	}
	sort.Slice(words, func(i, j int) bool {return freq[words[j]] < freq[words[i]]})
	for _, word := range words {
		fmt.Fprintf(os.Stderr, "%v\t%v\n", freq[word], word)
	}
*/
	fmt.Fprintf(os.Stdout, "counted morphemes occurrences total: %v\n", occurrencesTotal)
}
