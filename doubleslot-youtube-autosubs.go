package main
import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}

func main() {
	fillerLine := func() astisub.Line {
		const invisibleBegin, invisibleEnd = `{\1a&HFF&\3a&HFF&}`, `{\1a&H00&\3a&H7F&}`
		var filler string; flag.StringVar(&filler, "filler", invisibleBegin+"."+invisibleEnd, "")
		flag.Parse()
		return astisub.Line{Items: []astisub.LineItem{{Text: filler}}}
	} ()

	var err error
	panicCheck := func() {if err != nil {panic(err)}}

	inSubs, err := astisub.OpenFile(flag.Arg(0)); panicCheck()
	maxLinesPerSlot := make([]int, 2)
	for i,item := range inSubs.Items {
		m := &(maxLinesPerSlot[i%2])
		if len(item.Lines) > *m {*m = len(item.Lines)}
	}
	fmt.Fprintln(os.Stderr, maxLinesPerSlot)
	type switchPoint struct {j int; on bool}
	extractTime := func(sp switchPoint) time.Duration {
		item := inSubs.Items[sp.j]
		if sp.on {return item.StartAt}
		return item.EndAt
	}
	sps := make([]switchPoint, 0, len(inSubs.Items)*2)
	for j := range inSubs.Items {
		for _, on := range [...]bool{true, false} {
			sps = append(sps, switchPoint{j:j, on:on})
		}
	}
	if len(sps)!=cap(sps) {panic(fmt.Sprintf("%v!=%v", len(sps), cap(sps)))}
	sort.SliceStable(sps, func(i0, i1 int) bool {
		t := [2]time.Duration{}
		for k, sp := range [...]switchPoint{sps[i0], sps[i1]} {
			t[k] = extractTime(sp)
		}
		return t[0] < t[1]
	})

	outSubs := astisub.NewSubtitles()
	activeItems := [...]int{-1, -1}
	empty := func() bool {for _,j:=range activeItems {if j!=-1 {return false}}; return true}
	toggle := func(requestedJ int) {
		for i,j := range activeItems {if j==requestedJ {activeItems[i]=-1; return}}
		for i,j := range activeItems {if j==-1 {activeItems[i]=requestedJ; return}}
		panic("requested new j, but activeItems are full")
	}
	for _, sp := range sps {
		t := extractTime(sp)
		if !empty() {
			last := len(outSubs.Items)-1
			if t > outSubs.Items[last].StartAt {
				outSubs.Items[last].EndAt = t
			} else {
				outSubs.Items = outSubs.Items[:last]
			}
		}
		toggle(sp.j)
		if !empty() {
			item := &astisub.Item{StartAt: t}
			for i,j := range activeItems {
				lines := func() []astisub.Line {
					if j==-1 {return nil}
					return inSubs.Items[j].Lines
				} ()
				item.Lines = append(item.Lines, lines...)
				linesToFill := maxLinesPerSlot[i] - len(lines)
				for k:=0; k<linesToFill; k++ {
					item.Lines = append(item.Lines, fillerLine)
				}
			}
			outSubs.Items = append(outSubs.Items, item)
		}
	}
	err = outSubs.Write(flag.Arg(1)); panicCheck()
}
