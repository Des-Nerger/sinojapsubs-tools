package main
import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	var (lineCircumfix string; useMaxLinesPerItem bool; maxOverlapAutofixes int)
	fillerLine, fillerDoubleLine, doubleLineCircumfix := func() (astisub.Line, astisub.Line, string) {
		const invisibleBegin, invisibleEnd = `{\1a&HFF&\3a&HFF&}`, `{\1a&H00&\3a&H7F&}`
		var filler string; flag.StringVar(&filler, "filler", invisibleBegin+"."+invisibleEnd, "")
		flag.StringVar(&lineCircumfix, "lineCircumfix", invisibleBegin+"｜"+invisibleEnd, "")
		flag.BoolVar(&useMaxLinesPerItem, "useMaxLinesPerItem", false, "")
		flag.IntVar(&maxOverlapAutofixes, "maxOverlapAutofixes", 0, "")
		flag.Parse(); maxOverlapAutofixes &^= -1<<(strconv.IntSize-1)
		const doubleSizeBegin, doubleSizeEnd = `{\fs28}`, `{\fs14}`
		return astisub.Line{Items: []astisub.LineItem{{Text: filler}}},
		       astisub.Line{Items: []astisub.LineItem{{Text: doubleSizeBegin + filler + doubleSizeEnd}}},
		       doubleSizeBegin + lineCircumfix + doubleSizeEnd
	} ()

	sign := func(i int)int{isntZero:=i!=0;return i>>(strconv.IntSize-1)|int(*(*byte)(Pointer(&isntZero)))}
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
				switch sign(len(line.Items) - 1) {
				case +1:
					fmt.Fprintf(os.Stderr, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
					for _, item := range line.Items {
						fmt.Fprintf(os.Stderr, "%q\n", item.Text)
					}
					fmt.Fprintln(os.Stderr)
					item.Lines = append(item.Lines[:m.count], item.Lines[m.count+1:]...)
					continue
				case 0:
					if line.Items[0].Text=="." {item.Lines[m.count] = fillerLine}
				}
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
	boolToInt := func(b bool) int {return int(*(*byte)(Pointer(&b)))}
	overlapFixesCount := 0
	toggle := func(sp switchPoint) {
		_, ok := activeItems[sp.i]
		if !ok {activeItems[sp.i] = sp.j + boolToInt(!sp.on && overlapFixesCount<maxOverlapAutofixes); return}
		if sp.on {
			if !(overlapFixesCount<maxOverlapAutofixes) { j:=activeItems[sp.i]
				panic(fmt.Sprintf("%v: %v overlaps %v\n%v\n-------\n%v\n\n",
					flag.Arg(sp.i), j+1, sp.j+1, inSubs[sp.i].Items[j], inSubs[sp.i].Items[sp.j])) }
			overlapFixesCount++
		}
		delete(activeItems, sp.i)
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
		toggle(sp)
		if len(activeItems) > 0 {
			item := &astisub.Item{StartAt: t}
			for i := range inSubs {
				j, ok := activeItems[i]
				lines := func() []astisub.Line {
					if !ok {return nil}
					return inSubs[i].Items[j].Lines
				} ()
				m := maxLinesPerItem[i]
				if i!=len(inSubs)-1 {
					item.Lines = append(item.Lines, lines...)
				} else {
					for k, line := range lines {
						c := func() string {
							if useMaxLinesPerItem && !m.isDouble(k) {return lineCircumfix}
							return doubleLineCircumfix
						} ()
						item.Lines = append(item.Lines, astisub.Line{Items: []astisub.LineItem{{Text: c+line.Items[0].Text+c}}})
					}
				}
				if len(inSubs) >= 2 {
					linesToFill := int(m.count) - len(lines)
					dontUseItAndIsntLastSub := !useMaxLinesPerItem && i!=len(inSubs)-1
					for k:=0; k<linesToFill; k++ {
						item.Lines = append(item.Lines, func() astisub.Line {
							if useMaxLinesPerItem && !m.isDouble(len(lines)+k) || dontUseItAndIsntLastSub {return fillerLine}
							return fillerDoubleLine
						} ())
					}
				}
			}
			outSubs.Items = append(outSubs.Items, item)
		}
	}
	if maxOverlapAutofixes!=0 {fmt.Fprintf(os.Stderr, "%v overlap fixes have been applied\n", overlapFixesCount)}
	err = outSubs.Write(flag.Arg(flag.NArg()-1)); panicCheck()
}
