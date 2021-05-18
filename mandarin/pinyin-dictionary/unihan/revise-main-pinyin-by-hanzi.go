//go run revise-main-pinyin-by-hanzi.go <main-pinyin-by-hanzi.tsv.txt all-pinyin-syllables-by-hanzi.txt phonetic_series.txt >revised-main-pinyin-by-hanzi.tsv.txt
package main
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	. "unsafe"
)

func panicCheck(err error) {if err != nil {panic(err)}}
func sign(i int) int {isntZero:=i!=0; return i>>(strconv.IntSize-1) | int(*(*byte)(Pointer(&isntZero)))}
func main() {
	type entry struct {all, series []string}
	m := make(map[string]entry)
	func() {
		file, err := os.Open(os.Args[1]); panicCheck(err); defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			f := strings.Split(scanner.Text(), "\t")
			e := m[f[0]]
			if len(e.all) != 0 {panic(fmt.Sprintf("%#v all already exist", f[0]))}
			e.all = f[1:]
			m[f[0]] = e
		}
	} ()
	func() {
		file, err := os.Open(os.Args[2]); panicCheck(err); defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			f := strings.Split(scanner.Text(), " ")
			e := m[f[0]]
			if len(e.series) != 0 {panic(fmt.Sprintf("%#v series already exist", f[0]))}
			e.series = f[1:]
			m[f[0]] = e
		}
	} ()
	scanner, bw := bufio.NewScanner(os.Stdin), bufio.NewWriter(os.Stdout); defer bw.Flush()
	for scanner.Scan() {
		f := strings.SplitN(scanner.Text(), "\t", 2); e := m[f[0]]
		bag:=make(map[string]int,len(e.all)); for _,s:=range e.all{bag[s]=1}
		type Max struct{i int; s string}; max := Max{i:1}
		for _, s := range e.series {
			for _, s := range m[s].all {
				if i:=bag[s]; i>=1 {
					i++
					switch sign(i-max.i) {
					case +1: max=Max{i:i,s:s}
					case 0: max.s=""
					}
					bag[s]=i
				}
			}
		}
		bw.WriteString(f[0]); bw.WriteByte('\t')
		bw.WriteString(func() string {if max.s=="" {return f[1]}; return max.s} ())
		bw.WriteByte('\n')
	}
}
