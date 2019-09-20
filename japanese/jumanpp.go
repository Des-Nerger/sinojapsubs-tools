package main
import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Jumanpp struct {
	Path string
	cmd *exec.Cmd
	inPipe io.WriteCloser
	outPipe io.ReadCloser
}

func (j *Jumanpp) Start() {
	j.cmd = exec.Command(func() string {if j.Path==""{return "jumanpp"}; return j.Path}())
	j.inPipe, _ = j.cmd.StdinPipe()
	j.outPipe, _ = j.cmd.StdoutPipe()
	j.cmd.Start()
}

func (j *Jumanpp) Wait() {
	j.inPipe.Close()
	j.cmd.Wait()
}

/*
var (
	noUnescapedSpaces = regexp.MustCompile(`(\\.|[^ \\])*`)
	unescapeSpaces = strings.NewReplacer(`\ `, ` `)
)
*/

type Morpheme [5]string

func trimSuffixだIfItIsAppropriateAdjective(fields []string) {
	switch  {
	case strings.HasPrefix(fields[7], "ナ形容詞"),
	     fields[7] == "ナノ形容詞":
		const expectedSuffix = "だ"
		if strings.HasSuffix(fields[2], expectedSuffix) {
			fields[2] = fields[2][:len(fields[2])-len(expectedSuffix)]
		} else {
			fmt.Fprintf(os.Stdout, "%v %q suffix is not %q\n",
				fields[7], fields[2], expectedSuffix)
		}
	}
}

func (j *Jumanpp) AnalyzeLine(line string) (morphemes []Morpheme) {
	for len(line)>=3 && line[:2] == "# " {
	/*
		fmt.Fprintf(os.Stdout, "skipping \"# \"-prefixed line: %q\n", line)
		return
	*/
		line = line[2:]
	}
	terminatedLine := make([]byte, len(line)+1)
	terminatedLine[copy(terminatedLine, line)] = '\n'
	j.inPipe.Write(terminatedLine)
	sc := bufio.NewScanner(j.outPipe)
	見出し語, 活用形 := "", ""
	for sc.Scan() {
		line := sc.Text()
		if line == "EOS" {break}
		if line == "# ERROR" {
			fmt.Fprintf(os.Stdout, "[# ERROR] with:\n%s\n", terminatedLine)
			continue
		}
		fields := func(n int) []string {
			if line[0] == ' ' {
				return append([]string{" ", " ", " "},
					strings.SplitN(line[len(`  \  \  `):], " ", n-3)...)
			}
			return strings.SplitN(line, " ", n)
		} (12)
		if fields[0]=="@" && fields[1]!="@" {
			trimSuffixだIfItIsAppropriateAdjective(fields[1:])
			if fields[3] != 見出し語 || fields[10] != 活用形 {
				fmt.Fprintf(os.Stdout, "[%q, %q] != [%q, %q]\n", 見出し語, 活用形, fields[3], fields[10])
			}
			continue
		}
		trimSuffixだIfItIsAppropriateAdjective(fields)
		morphemes = append(morphemes, Morpheme{fields[0], fields[2], fields[3], fields[7], fields[9]})
		見出し語, 活用形 = fields[2], fields[9]
	}
	return
}
