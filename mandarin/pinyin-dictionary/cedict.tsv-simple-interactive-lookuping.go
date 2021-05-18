package main
import (
	"bufio"
	"fmt"
	"os"
	"strings"
)
func main() {
	type entry struct {pinyin, definition string}
	entries := map[string][]*entry{}
	func () {
		file, err := os.Open(os.Args[1]); if err!=nil{panic(err)}; defer file.Close()
		for sc := bufio.NewScanner(file); sc.Scan(); {
			f := strings.SplitN(sc.Text(), "\t", 4)
			e := &entry{pinyin: f[2], definition: f[3]}
			for i:=0;; i++ {entries[f[i]] = append(entries[f[i]], e); if i==1 || f[0]==f[1] {break}}
		}
	} ()
	for sc := bufio.NewScanner(os.Stdin); sc.Scan(); {
		key:=sc.Text(); if key=="" {continue}
		for _,e := range entries[key] {fmt.Printf("%v　%v\n", e.pinyin, e.definition)}
	}
}
