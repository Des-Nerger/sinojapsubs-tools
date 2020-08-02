package main
import (
	"bufio"
	"flag"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	. "unsafe"

	"github.com/Des-Nerger/porter2"
)
func main() {
	blanks, sign := "", func(i int)int{isntZero:=i!=0;return i>>(strconv.IntSize-1)|int(*(*byte)(Pointer(&isntZero)))}
	max := func(a, b int) int {if a>b {return a}; return b}
	var circumfix string;flag.StringVar(&circumfix,"circumfix",`{\1a&HFF&\3a&HFF&}｜{\1a&H00&\3a&H7F&}`,"");flag.Parse()
	frequentIDStems:=make(map[string]struct{},flag.NArg());for _,a:=range flag.Args(){frequentIDStems[a]=struct{}{}}
	bw:=bufio.NewWriter(os.Stdout); defer bw.Flush()
	for sb,scanner:=(strings.Builder{}),bufio.NewScanner(os.Stdin); scanner.Scan(); {
		line:=scanner.Text(); sbTargetCap:=len(circumfix)+len(line)+len(circumfix)
		idStart:=-1; var runeCount int
	innerFor:
		for i:=0;; {
			var (r rune; size int)
			switch sign(i-len(line)) {
			case -1: r,size=utf8.DecodeRuneInString(line[i:]); if r==utf8.RuneError{panic("failed to decode rune")}
			case  0: if idStart!=-1 {r,size=utf8.RuneError,1; break}; fallthrough
			case +1: break innerFor
			}
			if idStart==-1 {if unicode.Is(unicode.Letter,r){idStart=i;runeCount=1} else if sb.Cap()!=0{sb.WriteRune(r)}
			} else {
				if unicode.In(r, unicode.Letter, unicode.Digit) {runeCount++
				} else {
					id := line[idStart:i]
					if _, ok := frequentIDStems[porter2.Stem(id)]; ok {
						if sb.Cap()==0 {sb.Grow(sbTargetCap); sb.WriteString(circumfix); sb.WriteString(line[:idStart])}
						if len(blanks)<runeCount {blanks=strings.Repeat(" "/*"-"*/, max(32, 2*runeCount))}
						sb.WriteString(blanks[:runeCount])
					} else if sb.Cap()!=0 {sb.WriteString(id)}
					if sb.Cap()!=0 && r!=utf8.RuneError {sb.WriteRune(r)}
					idStart=-1
				}
			}
			i += size
		}
		bw.WriteString(func() string {
			if sb.Cap()==0 {return line}
			sb.WriteString(circumfix); if sb.Cap()!=sbTargetCap{panic("unexpected sb capacity")}
			s:=sb.String(); sb.Reset(); return s
		} ()); bw.WriteByte('\n')
	}
}
