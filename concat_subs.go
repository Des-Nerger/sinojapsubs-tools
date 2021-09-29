package main
import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

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
	flag.StringVar(&outputFilename, "o", "/dev/stdout", "output filename")
	flag.Parse()

	subses, duration, itemsCount := make([]*astisub.Subtitles, 0, flag.NArg()), time.Duration(0), 0
	for _, a := range flag.Args() {
		if a[0]=='+' {
			parsedDuration, err := time.ParseDuration(a); fatalCheck(err)
			duration += func() time.Duration {
				if len(subses)==0 {return 0}
				subs := subses[len(subses)-1]
				return subs.Items[len(subs.Items)-1].EndAt
			} () + parsedDuration
			continue
		}
		subs, err := astisub.OpenFile(a); fatalCheck(err)
		if len(subs.Items)==0 {continue}
		if duration!=0 {for _, it := range subs.Items {it.StartAt+=duration; it.EndAt+=duration}}
		subses=append(subses,subs); itemsCount+=len(subs.Items)
	}
	items := make([]*astisub.Item, 0, itemsCount)
	for _,subs := range subses {items = append(items, subs.Items...)}
	if cap(items)!=len(items) {fatalCheck(errors.New("cap(items)!=len(items)"))}
	subses[0].Items = items

	err := subses[0].Write(outputFilename); fatalCheck(err)
}
