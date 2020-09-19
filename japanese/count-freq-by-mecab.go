package main
import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
	//"github.com/miiton/kanaconv"
	"github.com/shogo82148/go-mecab"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func main() {
	panicCheck := func(err error) {if err != nil {panic(err)}}
	freq := map[/*[2]*/string]uint{}
	var occurrencesTotal uint64
	func() {
		tagger, err := mecab.New(map[string]string{"eos-format": "\x00"}); panicCheck(err); defer tagger.Destroy()
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
					result, err := tagger.Parse(lineString); panicCheck(err)
					for _, word := range strings.FieldsFunc(result, func(r rune)bool{return r=='\n'}) {
						const sep = ",\t"
						if strings.HasPrefix(word, sep) {continue}
						fields := strings.FieldsFunc(word, func(r rune)bool{return strings.ContainsRune(sep, r)})
						/*r:=fields[len(fields)-2]; if r=="*"{continue}; r=kanaconv.KatakanaToHiragana(r)
						w, d, i := fields[0], fields[len(fields)-3], 0
						for ; i<len(w) && i<len(d) && w[i]==d[i]; i++ {}
						freq[[...]string{d, r[:len(r)-(len(w)-i)]+d[i:]}]++*/
						freq[fields[len(fields)-3]]++
						occurrencesTotal++
					}
				}
			}
		}
	} ()
	type pair struct{uint; s /*[2]*/string}
	freqPairs := make([]pair, len(freq))
	{
		i := 0
		for s, count := range freq {
			freqPairs[i] = pair{uint:count, s:s}
			i++
		}
	}
	sort.Slice(freqPairs, func(i, j int) bool {return freqPairs[j].uint < freqPairs[i].uint})
	for _, pair := range freqPairs {
		fmt.Fprintf(os.Stderr, "%v\t%v"/*+"\t%v"*/+"\n", pair.uint, pair.s/*[0], pair.s[1]*/)
	}
	fmt.Fprintf(os.Stdout, "counted morphemes occurrences total: %v\n", occurrencesTotal)
}
