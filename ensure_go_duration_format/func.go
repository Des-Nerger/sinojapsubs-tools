package ensure_go_duration_format
import (
	"fmt"
	"strings"
)

func EnsureGoDurationFormat(s string) string {
	sepIndices, atMs := make([]int, 0, 3), true
	for i:=len(s)-1; i>=0; {
		if atMs {
			switch s[i] {
			case ','/*,'.'*/:
				sepIndices = append(sepIndices, i)
				atMs = false
			case ':':
				sepIndices = append(sepIndices, len(s))
				atMs = false
				continue
			}
		} else {
			switch s[i] {
			case ':': sepIndices = append(sepIndices, i)
			}
		}
		i--
	}
	if len(sepIndices)==0 {return s}
	prevSepI, sb := -1, strings.Builder{}; sb.Grow(len(s)+1)
	for i:=len(sepIndices)-1; i>=0; i-- {
		sepI := sepIndices[i]
		sb.WriteString(s[prevSepI+1:sepI])
		sb.WriteByte(func() byte {
			switch i {
			case 0:
				if sepI < len(s) {
					sb.WriteByte('.')
					sb.WriteString(s[sepI+1:])
				}
				return 's'
			case 1: return 'm'
			case 2: return 'h'
			}
			panic(fmt.Sprintf("%#v resulted in len(sepIndices%v)==%v>3\n", s, sepIndices, len(sepIndices)))
		} ())
		prevSepI = sepI
	}
	if sb.Len()!=sb.Cap() {
		panic(fmt.Sprintf("unexpected input and output strings' lengths: len(%#v)==%v --> len(%#v)==%v\n",
			s, len(s), sb.String(), sb.Len()))
	}
	//fmt.Fprintf(os.Stderr, "%#v --> %#v\n", s, sb.String())
	return sb.String()
}
