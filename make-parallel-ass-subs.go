package main
import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	. "unsafe"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}


type MaxLinesPerItem struct {
	doubles uint64
	count byte
}
func (m *MaxLinesPerItem) inc(withDouble bool) {
	m.doubles |= uint64(*(*byte)(Pointer(&withDouble))) << uint64(m.count)
	m.count++
}
func max(a, b byte) byte {if a>b {return a}; return b}
func (m *MaxLinesPerItem) max(n MaxLinesPerItem) MaxLinesPerItem {
	return MaxLinesPerItem{count: max(m.count, n.count), doubles: m.doubles | n.doubles}
}
func (m *MaxLinesPerItem) isDouble(lineIndex int) bool {
	return ((1 << lineIndex) & m.doubles) != 0
}
func (m MaxLinesPerItem) String() string {
	sb := strings.Builder{}; sb.Grow(int(m.count))
	for i:=0; i<int(m.count); i++ {
		sb.WriteByte(func() byte {
			if m.isDouble(i) {return '2'}
			return '1'
		} ())
	}
	return sb.String()
}

func main() {
	const doubleSizeBegin, doubleSizeEnd = `{\fs32}`, `{\fs16}`
	fillerLine, fillerDoubleLine := func() (astisub.Line, astisub.Line) {
		var filler string
		flag.StringVar(&filler, "filler", ".", "")
		flag.Parse()
		return astisub.Line{Items: []astisub.LineItem{{Text: filler}}},
		       astisub.Line{Items: []astisub.LineItem{{Text: doubleSizeBegin + filler + doubleSizeEnd}}}
	} ()

	var err error
	panicCheck := func() {if err != nil {panic(err)}}

	inSubs := []*astisub.Subtitles{}
	maxLinesPerItem := []MaxLinesPerItem{}
	for i, inFilename := range flag.Args()[:flag.NArg()-1] {
		inSubs = append(inSubs, nil); maxLinesPerItem = append(maxLinesPerItem, MaxLinesPerItem{count: 0})
		inSubs[i], err = astisub.OpenFile(inFilename); panicCheck()
		for j:=0; j<len(inSubs[i].Items); {
			item := inSubs[i].Items[j]
			initialLen := len(item.Lines)
			m := MaxLinesPerItem{count: 0}
			for ; int(m.count)<len(item.Lines); {
				line := item.Lines[m.count]
				//line.Items = []astisub.LineItem{{Text: line.String()}}
				if len(line.Items) > 1 {
					fmt.Fprintf(os.Stderr, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
					for _, item := range line.Items {
						fmt.Fprintf(os.Stderr, "%q\n", item.Text)
					}
					fmt.Fprintln(os.Stderr)
					item.Lines = append(item.Lines[:m.count], item.Lines[m.count+1:]...)
					continue
				}
				//fmt.Printf("%#v\n", line.Items[0].Text)
				m.inc(strings.Contains(line.Items[0].Text, `\fs`))
			}
			if m.count==0 {
				inSubs[i].Items = append(inSubs[i].Items[:j], inSubs[i].Items[j+1:]...)
				continue
			}
			maxLinesPerItem[i] = m.max(maxLinesPerItem[i])
			if int(m.count)!=initialLen {
				fmt.Fprintf(os.Stderr, "%v ====>>>> %v\n[[[\n  %#v\n]]]\n\n", initialLen, m.count, item.Lines)
			}
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
		if j != requestedJ {panic(fmt.Sprintf("%v: %v overlaps %v\n%v\n-------\n%v\n\n",
			flag.Arg(i), j+1, requestedJ+1, inSubs[i].Items[j], inSubs[i].Items[requestedJ]))}
		delete(activeItems, i)
	}
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
				m := maxLinesPerItem[i]
				if i==0 {
					item.Lines = append(item.Lines, lines...)
				} else {
					for k, line := range lines {
						c := func() string {
							const (
								circumfix = `{\1a&HFF&\3a&HFF&}｜{\1a&H00&\3a&H7F&}`
								doubleCircumfix = doubleSizeBegin + circumfix + doubleSizeEnd
							)
							if m.isDouble(k) {return doubleCircumfix}
							return circumfix
						} ()
						item.Lines = append(item.Lines, astisub.Line{Items: []astisub.LineItem{{Text: c+line.Items[0].Text+c}}})
					}
				}
				linesToFill := int(m.count) - len(lines)
				for k:=0; k<linesToFill; k++ {
					item.Lines = append(item.Lines, func() astisub.Line {
						if m.isDouble(len(lines)+k) {return fillerDoubleLine}
						return fillerLine
					} ())
				}
			}
			outSubs.Items = append(outSubs.Items, item)
		}
	}
	err = outSubs.Write(flag.Arg(flag.NArg()-1)); panicCheck()
}
