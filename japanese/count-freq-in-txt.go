package main
import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	freq := map[string]uint{}
	var occurrencesTotal, inputLinesCount uint64
	func() {
		j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
		j.Start()
		defer j.Wait()
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			inputLinesCount++
			//if inputLinesCount < 29500000 {continue}
			if inputLinesCount & (128*1024 - 1) == 0 {
				fmt.Fprintln(os.Stdout, inputLinesCount)
			}
			for _, sentence := range strings.SplitAfter(sc.Text(), "。") {
				for _, morpheme := range j.AnalyzeLine(sentence) {
					switch morpheme[2] {
					case "特殊", "未定義語", "0": // Do nothing
					default:
						freq[morpheme[1]]++
						occurrencesTotal++
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
