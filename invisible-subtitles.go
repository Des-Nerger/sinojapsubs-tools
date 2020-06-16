package main
import (
	"bufio" //BW:
	"fmt"
	"os"
	//"runtime"
	"strconv"
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
	/*
		fmt.Fprintln(os.Stdout, err)
		runtime.Goexit()
	*/
		//os.Exit(1)
		panic(err)
	}
}

func main() {
	//defer os.Exit(0)
	inSubs, err := astisub.OpenFile(os.Args[1]); fatalCheck(err)
//*BW:
	bw := bufio.NewWriter(os.Stderr)
	const highlightColor = `#ffad00`
	bw.WriteString(`<head>
	<meta charset="utf-8"/>
	<style>
		::-moz-selection {background: #9ab87cfe;}
		::selection {background: #9ab87cfe;}
		body {
		  font-size: 24px;
		  color: #dcddde; background-color: #36393f;
		}
		span.i {
			color: #202225; background-color: #202225;
		}
		span.i::-moz-selection {color: #99b77b;}
		span.i::selection {color: #99b77b;}
		span.h {
			//background-color: #b58900;
			color: `+highlightColor+`;
		}
		span.i.h {
			color: #b58900;
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
//*/
	writtenItemsCount := 0
	outSubs := astisub.NewSubtitles()
	for _, inItem := range inSubs.Items {
		//containsMarkedOrUnhidden := false
		for _, line := range inItem.Lines {
			if len(line.Items) > 1 {
				fmt.Fprintf(os.Stdout, "len(line.Items)==%v; skipping all of them\n", len(line.Items))
				for _, inItem := range line.Items {
					fmt.Fprintf(os.Stdout, "%q\n", inItem.Text)
				}
				fmt.Fprintln(os.Stdout)
				continue
			}
		/*
			if strings.ContainsAny(line.Items[0].Text, "<|") {
				containsMarkedOrUnhidden = true
				break
			}
		*/
		}
		//if !containsMarkedOrUnhidden {continue}
		outItem := &astisub.Item{StartAt: inItem.StartAt, EndAt: inItem.EndAt}
		writtenItemsCount++
		writtenItemsCountString := strconv.Itoa(writtenItemsCount) //fmt.Sprintf("%03d", writtenItemsCount)
		fmt.Fprintln(bw, writtenItemsCountString) //BW:
		outItem.Lines = append(outItem.Lines, astisub.Line{Items: []astisub.LineItem{{Text: writtenItemsCountString}}})
		for _, line := range inItem.Lines {
			if len(line.Items) > 1 {continue}
			lineText := line.Items[0].Text
			if strings.HasPrefix(lineText, "#") {
				fmt.Fprintln(bw, lineText) //BW:
				outItem.Lines = append(outItem.Lines, astisub.Line{Items: []astisub.LineItem{{Text: lineText}}})
				continue
			}
			lineBuilder := strings.Builder{}; lineBuilder.Grow(len(lineText))
			const (
				invisibleSpanStart = `<span class=i>`
				visibleHighlightedSpanStart = `<span class=h>`
				invisibleHighlightedSpanStart = `<span class="i h">`
				spanEnd = `</span>`
				fontEnd = `</font>`
			)
			currentSpanStart := ""
			inAngleBrackets := false
			hideRunes := false
			for _, r := range lineText {
				if r=='|' {hideRunes = !hideRunes; continue}
				hideRune := hideRunes && unicode.In(r, unicode.Han, /*unicode.Hiragana, unicode.Katakana,*/ unicode.Bopomofo, )
				if inAngleBrackets {
					if r=='>' {inAngleBrackets = false; continue}
					switch currentSpanStart {
					case "":
						currentSpanStart = func() string {
							if hideRune {return invisibleHighlightedSpanStart}
							return visibleHighlightedSpanStart
						} ()
						bw.WriteString(currentSpanStart) //BW:
						lineBuilder.WriteString(`<font color="` + highlightColor + `">`)
					case invisibleSpanStart:
						bw.WriteString(spanEnd) //BW:
						currentSpanStart = func() string {
							if hideRune {return invisibleHighlightedSpanStart}
							return visibleHighlightedSpanStart
						} ()
						bw.WriteString(currentSpanStart) //BW:
					case visibleHighlightedSpanStart:
						if hideRune {
							bw.WriteString(spanEnd) //BW:
							currentSpanStart = invisibleHighlightedSpanStart
							bw.WriteString(currentSpanStart) //BW:
						}
					case invisibleHighlightedSpanStart:
						if !hideRune {
							bw.WriteString(spanEnd) //BW:
							currentSpanStart = visibleHighlightedSpanStart
							bw.WriteString(currentSpanStart) //BW:
						}
					}
				} else {
					if r=='<' {inAngleBrackets = true; continue} //BW:
					switch currentSpanStart {
					case "":
						if hideRune {
							currentSpanStart = invisibleSpanStart
							bw.WriteString(currentSpanStart) //BW:
						}
					case invisibleSpanStart:
						if !hideRune {
							bw.WriteString(spanEnd) //BW:
							currentSpanStart = ""
						}
					case visibleHighlightedSpanStart, invisibleHighlightedSpanStart:
						bw.WriteString(spanEnd) //BW:
						lineBuilder.WriteString(fontEnd)
						currentSpanStart = func() string {
							if hideRune {return invisibleSpanStart}
							return ""
						} ()
						bw.WriteString(currentSpanStart) //BW:
					}
				}
				bw.WriteRune(r) //BW:
				lineBuilder.WriteRune(func() rune {
					switch currentSpanStart {
					case invisibleSpanStart: return '＿' // '■' '　' '￢' '一' '＝'
					case invisibleHighlightedSpanStart: return '口'
					default: fallthrough; case "", visibleHighlightedSpanStart: return r
					}
				} ())
			}
			if inAngleBrackets {
				fmt.Fprintf(os.Stdout, "unterminated '<' in %q\n", lineText)
			}
			if currentSpanStart != "" {
				bw.WriteString(spanEnd) //BW:
				lineBuilder.WriteString(fontEnd)
			}
			fmt.Fprintln(bw) //BW:
			outItem.Lines = append(outItem.Lines, astisub.Line{Items: []astisub.LineItem{{Text: lineBuilder.String()}}})
		}
		fmt.Fprintln(bw) //BW:
		outSubs.Items = append(outSubs.Items, outItem)
	}
	if len(os.Args) >= 3 {
		err = outSubs.Write(os.Args[2]); fatalCheck(err)
	}
}
