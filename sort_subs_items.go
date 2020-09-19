package main
import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"

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

	inputFilename := flag.Arg(0)
	subs, err := astisub.OpenFile(inputFilename); fatalCheck(err)
	sort.SliceStable(subs.Items, func(i,j int) bool {return subs.Items[i].StartAt < subs.Items[j].StartAt})
	if outputFilename == "" {
		outputFilename = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFilename, "$2.sorted.srt")
		if inputFilename == outputFilename {
			fatalCheck(errors.New("inputFilename == outputFilename"))
		}
	}
	err = subs.Write(outputFilename); fatalCheck(err)
}
