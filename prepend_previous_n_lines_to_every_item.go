package main
import (
	"flag"
	//"fmt"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}

func max(a, b int) int {if a > b {return a}; return b}

func main() {
	var (n int; err error; removeItemOwnLines bool)
	separatorLine := func() astisub.Line {
		flag.IntVar(&n, "n", 6, "")
		var separator string; flag.StringVar(&separator, "separator", "|", "")
		flag.BoolVar(&removeItemOwnLines, "removeItemOwnLines", false, "")
		flag.Parse()
		return astisub.Line{Items: []astisub.LineItem{{Text: separator}}}
	} ()
	panicCheck := func() {if err != nil {panic(err)}}
	subs, err := astisub.OpenFile(flag.Arg(0)); panicCheck()
	linesCount := 0; for _, item := range subs.Items {linesCount += 1 + len(item.Lines)}
	lines := make([]astisub.Line, 0, linesCount)
	for _, item := range subs.Items {
		lines = append(append(lines, separatorLine), item.Lines...)
		item.Lines = lines[max(0, len(lines)-(len(item.Lines)+n)) : len(lines) - func() int {
			if !removeItemOwnLines {return 0}
			return len(item.Lines)//+1
		} ()]
	}
	if len(lines) != cap(lines) {panic(nil)}
	//if len(subs.Items)>=1 && removeItemOwnLines {subs.Items = subs.Items[1:]}
	err = subs.Write(flag.Arg(1)); panicCheck()
}
