package main
import (
	"os"
	//"unicode"

	"github.com/Des-Nerger/go-astisub"
	"github.com/Des-Nerger/sinojapsubs-tools/rubytagpolyfill"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}

func main() {
/*
	containsOnly := func(s string, ranges ...*unicode.RangeTable) bool {
		for _, r := range s {if !unicode.In(r, ranges...) {return false}}
		return true
	}
	ascii := &unicode.RangeTable{R16: []unicode.Range16{{0, unicode.MaxASCII, 1}}, LatinOffset: 1}
*/
	var err error
	panicCheck := func() {if err != nil {panic(err)}}
	subs, err := astisub.OpenFile(os.Args[1]); panicCheck()
	for _, item := range subs.Items {
		newLines := make([]astisub.Line, 0, len(item.Lines))
		for _, line := range item.Lines {
			lineText := line.Items[0].Text
			//if containsOnly(lineText, ascii, unicode.Cyrillic) {newLines = append(newLines, line); continue}
			newLineTexts := rubytagpolyfill.ToSpacefilledTwoLines(lineText, `{\fs30}`, `{\fs15}`)
		/*
			circumfix := func() string {
				if len(newLineTexts)<=1 {return ""}
				return `{\1a&H7F&\3a&H7F&}｜{\1a&H00&\3a&H7F&}`
			} ()
		*/
			for _, newLineText := range newLineTexts {
				newLines = append(newLines,
					astisub.Line{Items: []astisub.LineItem{{Text: /*circumfix +*/ newLineText /*+ circumfix*/}}})
			}
		}
		item.Lines = newLines
	}
	err = subs.Write(os.Args[2]); panicCheck()
}
