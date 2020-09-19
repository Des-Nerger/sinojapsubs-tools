package main
import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"

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

func main() {
	defer os.Exit(0)

	var outputFilename string
	flag.StringVar(&outputFilename, "o", "", "output filename")
	flag.Parse()

	inputFilename, re := flag.Arg(0), regexp.MustCompile(flag.Arg(1))
	subs, err := astisub.OpenFile(inputFilename); fatalCheck(err)
	newItems := make([]*astisub.Item, 0, len(subs.Items))
	for _, it := range subs.Items {
		newLines := make([]astisub.Line, 0, len(it.Lines))
		for _, l := range it.Lines {
			switch {
			case len(l.Items)!=1:
			case re.MatchString(l.Items[0].Text): continue
			}
			newLines = append(newLines, l)
		}
		if len(newLines) >= 1 {
			it.Lines = newLines
			newItems = append(newItems, it)
		}
	}
	subs.Items = newItems
	if outputFilename == "" {
		outputFilename = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFilename, "$2.lines-deleted.srt")
		if inputFilename == outputFilename {
			fatalCheck(errors.New("inputFilename == outputFilename"))
		}
	}
	err = subs.Write(outputFilename); fatalCheck(err)
}
