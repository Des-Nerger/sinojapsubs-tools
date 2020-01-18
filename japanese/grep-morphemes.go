package main
import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {
	astilog.SetLogger(astilog.New( astilog.Configuration{Verbose: true} ))
}

func main() {
	re := regexp.MustCompile(os.Args[1])
	j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
	j.Start()
	defer j.Wait()
	for _, arg := range os.Args[2:] {
		subs, err := astisub.OpenFile(arg)
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
			continue
		}
		for _, item := range subs.Items {
			for _, line := range item.Lines {
				lineString := func() string {
					if len(line.Items) > 1 {
						var texts []string
						for _, lineItem := range line.Items {
							texts = append(texts, lineItem.Text)
						}
						return strings.Join(texts, "")
					}
					return line.Items[0].Text
				} ()
				for _, morpheme := range j.AnalyzeLine(lineString) {
					switch morpheme[2] {
					case "特殊", "未定義語", "0": // Do nothing
					default:
						if re.MatchString(morpheme[1]) {
							fmt.Fprintf(os.Stdout, "%v:\n%v\n\n", arg, lineString)
						}
					}
				}
			}
		}
	}
}
