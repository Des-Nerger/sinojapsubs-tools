package main
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	syllables := map[string]struct{}{}
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "\t")
		if len(s)>3 {panic(fmt.Sprintf(">3 fields: %#v", s))}
		if len(s)<3 || s[1]!="kHanyuPinyin" {continue}
		for _, e := range strings.Split(s[2], " ") {
			for _, syllable := range strings.Split(e[strings.IndexByte(e, ':')+1:], ",") {
				syllables[syllable] = struct{}{}
			}
		}
	}
	for syllable := range syllables {
		bw.WriteString(syllable)
		bw.WriteByte('\n')
	}
}
