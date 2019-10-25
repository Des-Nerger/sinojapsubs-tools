package main
import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"unicode"

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
	inSubs, err := astisub.OpenFile(os.Args[1]); fatalCheck(err)
	bw := bufio.NewWriter(os.Stderr)
	bw.WriteString(`<head>
	<meta charset="utf-8"/>
	<style>
		::-moz-selection {
			/* color: #9ab87c; */
			color: #99b77b;
			background: #9ab87cfe;
		}
		::selection {
			/* color: #9ab87c; background: #9ab87cfe; */
			color: #99b77b; background: #9ab87cfe;
			/* color: #0a66d8; background: #0b67d9fe; */
		}
		body {
		  font-size: 20px;
		  /* background-color: #f4f4f4; */
		  color: #dcddde; background-color: #36393f;
		}
		span {
			/* color: #d0d0d0; background-color: #d0d0d0; */
			color: #202225; background-color: #202225;
		}
		span.h {
			/* color: #ffff00; background-color: #ffff00; */
			color: #b58900; background-color: #b58900;
		}
		span.h::-moz-selection {
			color: #9ab87b;
		}
		span.h::selection {
			color: #9ab87b;
			/* color: #0b67d8; */
		}
	</style>
</head>
<body>
<pre>
`)
	defer func() {
		bw.WriteString(`























</pre>
</body>
`)
		bw.Flush()
	} ()
	writtenItemsCount := 0
	outSubs := astisub.NewSubtitles()
	for _, inItem := range inSubs.Items {
		containsMarked := false
		for _, line := range inItem.Lines {
			if len(line.Items) > 1 {
				fmt.Fprintf(os.Stdout, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
				for _, inItem := range line.Items {
					fmt.Fprintf(os.Stdout, "%q\n", inItem.Text)
				}
				fmt.Fprintln(os.Stdout)
				continue
			}
			if strings.ContainsRune(line.Items[0].Text, '<') {
				containsMarked = true
				break
			}
		}
		if !containsMarked {continue}
		outItem := &astisub.Item{StartAt: inItem.StartAt, EndAt: inItem.EndAt}
		writtenItemsCount++
		writtenItemsCountString := fmt.Sprintf("%03d", writtenItemsCount)
		fmt.Fprintln(bw, writtenItemsCountString)
		outItem.Lines = append(outItem.Lines, astisub.Line{Items: []astisub.LineItem{{Text: writtenItemsCountString}}})
		for _, line := range inItem.Lines {
			if len(line.Items) > 1 {continue}
			const (
				visible = iota
				invisible
				invisibleHighlighted
				inAngleBrackets
			)
			state := visible
			const (
				invisibleSpanStart = "<span>"
				invisibleHighlightedSpanStart = "<span class=h>"
				spanEnd = "</span>"
			)
			shouldBeHidden := func (r rune) bool {
				return unicode.In(r, /*unicode.Letter, unicode.Number,*/ unicode.Han, unicode.Hiragana, unicode.Katakana, )
			}
			lineBuilder := strings.Builder{}; lineBuilder.Grow(len(line.Items[0].Text))
			for _, r := range line.Items[0].Text {
				switch state {
				case visible:
					if r == '<' {
						state = inAngleBrackets
						bw.WriteString(invisibleHighlightedSpanStart)
						continue
					}
					if shouldBeHidden(r) {
						state = invisible
						bw.WriteString(invisibleSpanStart)
					}
				case invisible:
					if r == '<' {
						state = inAngleBrackets
						bw.WriteString(spanEnd)
						bw.WriteString(invisibleHighlightedSpanStart)
						continue
					}
					if !shouldBeHidden(r) {
						state = visible
						bw.WriteString(spanEnd)
					}
				case invisibleHighlighted:
					if r == '<' {
						state = inAngleBrackets
						continue
					}
					bw.WriteString(spanEnd)
					if shouldBeHidden(r) {
						state = invisible
						bw.WriteString(invisibleSpanStart)
					} else {
						state = visible
					}
				case inAngleBrackets:
					if r == '>' {
						state = invisibleHighlighted
						continue
					}
				}
				bw.WriteRune(r)
				lineBuilder.WriteRune(func() rune {
					switch state {
					case visible: return r
					case invisible: return '一'
					case inAngleBrackets: return '口'
					default: fallthrough; case invisibleHighlighted:
						panic("unexpected state: " + string(state))
					}
				} ())
			}
			switch state {
			case inAngleBrackets:
				fmt.Fprintf(os.Stdout, "unterminated '<' in %q\n", line.Items[0].Text)
				fallthrough
			case invisible, invisibleHighlighted:
				bw.WriteString(spanEnd)
			}
			fmt.Fprintln(bw)
			outItem.Lines = append(outItem.Lines, astisub.Line{Items: []astisub.LineItem{{Text: lineBuilder.String()}}})
		}
		fmt.Fprintln(bw)
		outSubs.Items = append(outSubs.Items, outItem)
	}
	err = outSubs.Write(os.Args[2]); fatalCheck(err)
}
