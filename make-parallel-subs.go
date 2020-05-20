package main
import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}

func main() {
	var err error
	panicCheck := func() {if err != nil {panic(err)}}

	inSubs := []*astisub.Subtitles{}
	maxLinesPerItem := []int{}
	for i, inFilename := range os.Args[1:len(os.Args)-1] {
		inSubs = append(inSubs, nil); maxLinesPerItem = append(maxLinesPerItem, 0)
		inSubs[i], err = astisub.OpenFile(inFilename); panicCheck()
		for j:=0; j<len(inSubs[i].Items); {
			item := inSubs[i].Items[j]
			initialLen := len(item.Lines)
			k:=0
			for ; k<len(item.Lines); {
				line := item.Lines[k]
				if len(line.Items) > 1 {
					fmt.Fprintf(os.Stderr, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
					for _, item := range line.Items {
						fmt.Fprintf(os.Stderr, "%q\n", item.Text)
					}
					fmt.Fprintln(os.Stderr)
					item.Lines = append(item.Lines[:k], item.Lines[k+1:]...)
					continue
				}
				k++
			}
			if k==0 {
				inSubs[i].Items = append(inSubs[i].Items[:j], inSubs[i].Items[j+1:]...)
				continue
			}
			if len(item.Lines) > maxLinesPerItem[i] {maxLinesPerItem[i] = len(item.Lines)}
			if k!=initialLen {fmt.Fprintf(os.Stderr, "%v ====>>>> %v\n[[[\n  %#v\n]]]\n\n", initialLen, k, item.Lines)}
			j++
		}
	}
	fmt.Fprintln(os.Stderr, maxLinesPerItem)
	type switchPoint struct {
		i, j int
		on bool
	}
	extractTime := func(sp switchPoint) time.Duration {
		item := inSubs[sp.i].Items[sp.j]
		if sp.on {return item.StartAt}
		return item.EndAt
	}
	sps := []switchPoint{}
	for i := range inSubs {
		for j := range inSubs[i].Items {
			for _, on := range [...]bool{true, false} {
				sps = append(sps, switchPoint{i:i, j:j, on:on})
			}
		}
	}
	sort.SliceStable(sps, func(i0, i1 int) bool {
		t := [2]time.Duration{}
		for k, sp := range [...]switchPoint{sps[i0], sps[i1]} {
			t[k] = extractTime(sp)
		}
		return t[0] < t[1]
	})

	outSubs := astisub.NewSubtitles()
	activeItems := make(map[int]int, len(inSubs))
	toggle := func(i, requestedJ int) {
		j, ok := activeItems[i]
		if !ok {activeItems[i] = requestedJ; return}
		if j != requestedJ {panic(fmt.Sprintf("%v: %v overlaps %v", os.Args[1+i], j+1, requestedJ+1))}
		delete(activeItems, i)
	}
	fillerLine := astisub.Line{Items: []astisub.LineItem{{Text: "."}}}
	for _, sp := range sps {
		t := extractTime(sp)
		if len(activeItems) > 0 {
			last := len(outSubs.Items)-1
			if t > outSubs.Items[last].StartAt {
				outSubs.Items[last].EndAt = t
			} else {
				outSubs.Items = outSubs.Items[:last]
			}
		}
		toggle(sp.i, sp.j)
		if len(activeItems) > 0 {
			item := &astisub.Item{StartAt: t}
			for i := range inSubs {
				j, ok := activeItems[i]
				lines := func() []astisub.Line {
					if !ok {return nil}
					return inSubs[i].Items[j].Lines
				} ()
				item.Lines = append(item.Lines, lines...)
				linesToFill := maxLinesPerItem[i] - len(lines)
				for k:=0; k<linesToFill; k++ {
					item.Lines = append(item.Lines, fillerLine)
				}
			}
			outSubs.Items = append(outSubs.Items, item)
		}
	}
	err = outSubs.Write(os.Args[len(os.Args)-1]); panicCheck()
}
