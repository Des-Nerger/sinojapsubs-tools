package main
import (
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astisub"
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
	//fmt.Printf("%#v\n", flag.Args())

	inputFilename := flag.Args()[0]
	subs, err := astisub.OpenFile(inputFilename); fatalCheck(err)
	ds := []time.Duration{}
	for _, arg := range flag.Args()[1:] {
		duration, err := time.ParseDuration(arg); fatalCheck(err)
		ds = append(ds, duration)
	}
	subs.Sync(ds...)
	if outputFilename == "" {
		outputFilename = regexp.MustCompile(`^(.*/|)([^/]+)\.[^.]+$`).
			ReplaceAllString(inputFilename, "$2.synced.srt")
		if inputFilename == outputFilename {
			fatalCheck(errors.New("inputFilename == outputFilename"))
		}
	}
	err = subs.Write(outputFilename); fatalCheck(err)
}
