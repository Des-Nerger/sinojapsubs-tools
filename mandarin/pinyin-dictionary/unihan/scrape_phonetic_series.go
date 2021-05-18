//awk -F'\t' 'NF>=3{printf$1" "} END{print""}' all-pinyin-syllables-by-hanzi.txt | xargs go run scrape_phonetic_series.go >phonetic_series.txt
package main
import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	. "unsafe"
)

func bytesToString(b []byte) string {return *(*string)(Pointer(&b))}
func main() {
	panicCheck:=func(e error){if e!=nil{panic(e)}}
	cl := &http.Client{Transport: &http.Transport{}}
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()
	for i, hanzi := range os.Args[1:] {func(){
		if i%512==0 {fmt.Fprintln(os.Stderr, i)}
		//_,size:=utf8.DecodeRuneInString(hanzis); hanzi:=hanzis[:size]; hanzis=hanzis[size:]
		const prefix="http://localhost:8080/wiktionary_en_all_2017-03/A/"
		rs,e:=cl.Get(prefix+hanzi+".html"); panicCheck(e); defer rs.Body.Close()
		b,e:=ioutil.ReadAll(rs.Body); panicCheck(e)
		d:=xml.NewDecoder(bytes.NewReader(b)); d.Strict=false; d.AutoClose=xml.HTMLAutoClose; d.Entity=xml.HTMLEntity
		m:=0
	innerFor:
		for {
			t, e := d.RawToken(); switch e{case nil:; case io.EOF:break innerFor; default:panic(e)}
			switch t := t.(type) {
			case xml.StartElement:
				switch {
				case m==0 && t.Name.Local=="tbody": m=1
				case m==3 && len(t.Attr)>=1: for _,a:=range t.Attr{if a.Value=="Hani"{m=4; break}}
				}
			case xml.EndElement:switch{case m>=1&&t.Name.Local=="tbody":if m>=3{bw.WriteByte('\n');m=-1}else{m=0}}
			case xml.CharData: s:=bytesToString(t)
				switch m {
				case 1: if strings.Contains(s,"phonetic series") {m=2}
				case 2: if strings.Contains(s,"Old Chinese") {m=3; bw.WriteString(hanzi)}
				case 4: if s!=hanzi {bw.WriteByte(' '); bw.WriteString(s)}; m=3
				}
			}
		}
	}()}
}
