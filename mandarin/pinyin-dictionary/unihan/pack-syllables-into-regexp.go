package main
import (
	"bufio"
	//"fmt"
	"os"
	//"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var removeDiacriticsTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
func removeDiacritics(s string) string  {
	output, _, e := transform.String(removeDiacriticsTransformer, s)
	if e != nil {
		panic(e)
	}
	return output
}

func main() {
/*
	curRemovedRuneIndex, removedRuneIndices := 0, []int{}
	isMn := func(r rune) bool {
		if unicode.Is(unicode.Mn, r) // Mn: nonspacing marks {
			removedRuneIndices = append(removedRuneIndices, curRemovedRuneIndex)
			curRemovedRuneIndex++
			return true
		}
		curRemovedRuneIndex++
		return false
	}
	removeDiacriticsTransformer := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	removeDiacritics := func(s string) string {
		s, _, e := transform.String(removeAccentsTransformer, s)
		if e != nil {panic(e)}
		curRemovedRuneIndex = 0, removedRuneIndices[:0]
		return s
	} 
*/

	scanner := bufio.NewScanner(os.Stdin)
	bw := bufio.NewWriter(os.Stdout); defer bw.Flush()

	groupedSyllables := make([][]rune, 0, 5)
	differingRuneIndex := -1
	writeGroupedSyllables := func () {
		switch len(groupedSyllables) {
		case 0: return
		case 1:
			bw.WriteString(string(groupedSyllables[0]))
		default:
			bw.WriteString(string(groupedSyllables[0][:differingRuneIndex]))
			bw.WriteByte('[')
			for _, syllable := range groupedSyllables {
				bw.WriteRune(syllable[differingRuneIndex])
			}
			bw.WriteByte(']')
			bw.WriteString(string(groupedSyllables[0][differingRuneIndex+1:]))
		}
		bw.WriteByte('\n')
	}
	for scanner.Scan() {
		prevSyllable := func() []rune {
			if len(groupedSyllables)==0 {return nil}
			return groupedSyllables[len(groupedSyllables)-1]
		} ()
		curSyllable := []rune(scanner.Text())
		result := differOnlyInDiacritics(prevSyllable, curSyllable)
		if result >= 0 {
			differingRuneIndex = result
			groupedSyllables = append(groupedSyllables, curSyllable)
			continue
		}
		writeGroupedSyllables()
		differingRuneIndex = result
		groupedSyllables = groupedSyllables[:0]
		groupedSyllables = append(groupedSyllables, curSyllable)
	}
	writeGroupedSyllables()
}

func differOnlyInDiacritics(a, b []rune) int {
	if len(a)!=len(b) {return -1}
	lastDifferingIndex := -1
	for i := range a {
		if a[i] == b[i] {continue}
		if removeDiacritics(string(a[i])) != removeDiacritics(string(b[i])) {return -1}
		lastDifferingIndex = i
	}
	return lastDifferingIndex
}

/*
func differIn1Rune(a, b []rune) int {
	if len(a)!=len(b) {return -1}
	differingIndex := -1
	for i := range a {
		if a[i] == b[i] {continue}
		if differingIndex>=0 {return -1}
		differingIndex = i
	}
	return differingIndex
}
*/
