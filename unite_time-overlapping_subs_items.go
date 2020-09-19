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

	inputFilename := flag.Args()[0]
	subs, err := astisub.OpenFile(inputFilename); fatalCheck(err)
	newItems := make([]*astisub.Item, 0, len(subs.Items))
	for _, it := range subs.Items {
		if len(newItems)==0 {
			newItems=append(newItems, it)
			continue
		}
		lastIt := newItems[len(newItems)-1]
		if lastIt.EndAt <= it.StartAt {
			newItems=append(newItems, it)
		} else {
			lastIt.EndAt = it.EndAt
			lastIt.Lines = append(lastIt.Lines, it.Lines...)
		}
	}
	subs.Items = newItems
	if outputFilename == "" {
		outputFilename = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFilename, "$2.united.srt")
		if inputFilename == outputFilename {
			fatalCheck(errors.New("inputFilename == outputFilename"))
		}
	}
	err = subs.Write(outputFilename); fatalCheck(err)
}
