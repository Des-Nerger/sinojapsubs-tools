package main
import (
	"bufio"
	//"fmt"
	"os"
	"strings"
	//"unicode"
	//"unicode/utf8"

	"github.com/hermanschaaf/cedict"
	"golang.org/x/text/unicode/norm"
)

func main() {
	c := cedict.New(os.Stdin)
	wc := norm.NFC.Writer(os.Stdout)
	bw := bufio.NewWriter(wc)
	defer func() {
		bw.Flush()
		wc.Close()
	} ()
	for {
		err := c.NextEntry()
		if err != nil {
			break
		}
		entry := c.Entry()
		bw.WriteString(entry.Traditional); bw.WriteByte('\t')
		bw.WriteString(entry.Simplified); bw.WriteByte('\t')
		bw.WriteString(strings.ReplaceAll(pinyin(strings.ToLower(entry.Pinyin)), " ", ".")); bw.WriteByte('\t')
		bw.WriteString(strings.Join(entry.Definitions, "/")); bw.WriteByte('\n')
	}
}
