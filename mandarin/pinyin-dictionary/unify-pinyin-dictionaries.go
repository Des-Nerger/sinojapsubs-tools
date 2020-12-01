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
	alreadyOccured := map[string]struct{}{}
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "\t")
		if len(s)!=2 {panic(fmt.Sprintf("len(%#v)!=2", s))}
		if _, ok := alreadyOccured[s[0]]; ok {continue}
		alreadyOccured[s[0]] = struct{}{}
		bw.WriteString(s[0])
		bw.WriteByte('\t')
		bw.WriteString(strings.Replace(s[1], " ", "." /*"'" "·"*/, -1))
		bw.WriteByte('\n')
	}
}
