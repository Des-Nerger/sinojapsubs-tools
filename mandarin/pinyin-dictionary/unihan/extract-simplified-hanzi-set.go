package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	count := 0
	for scanner.Scan() {
		splits := strings.Split(scanner.Text(), "\t")
		if len(splits)>3 {panic(fmt.Sprintf(">3 fields: %#v", splits))}
		if len(splits)<3 || splits[1]!="kSimplifiedVariant" || splits[2]=="" {continue}
		for _, code := range strings.Split(splits[2], " ") {
			if !strings.HasPrefix(code, "U+") {panic("unexpected prefix: " + code)}
			i, err := strconv.ParseInt(code[2:], 16, 32)
			if err!=nil {panic(err)}
			r := rune(i)
			if !unicode.In(r, unicode.Han) {
				panic(fmt.Sprintf("not a Han character: %v", code))
			}
			bw.WriteRune(r)
			count++
		}
	}
	os.Stderr.WriteString(strconv.Itoa(count) + "\n")
}
