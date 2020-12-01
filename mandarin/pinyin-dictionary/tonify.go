package main
import (
	"fmt"; "os"
	"regexp"
	"strings"
)

var utones = map[string]string{"1":"\u0304", "2":"\u0301", "3":"\u030C", "4":"\u0300", "5":""}
func tonify(vowels, tone string) string {
	if vowels == "ou" {return "o" + utones[tone] + "u"}
	tonified := false
	text := strings.Builder{}
	for i, r := range vowels {
		text.WriteRune(r)
		if r=='a' || r=='e' || i==len(vowels)-1 && !tonified {
			text.WriteString(utones[tone])
			tonified = true
		}
	}
	return strings.Replace(text.String(), "u:", "\u00FC", 1)
}

var syllableRegexp = regexp.MustCompile(`([^aeiou:]*)([aeiou:]+)([^aeiou:]*)([1-5])`)
func pinyin(syllables string) string {
	text := strings.Builder{}
	for i, syllable := range strings.Fields(syllables) {
		if i>0 {
			text.WriteByte(' ')
		}
		switch syllable {
		case "r5", "m2", "m4", "n2", "n3", "n4", "ng2", "ng3", "ng4", "hng5", "xx5":
			text.WriteString(syllable[:1])
			last := len(syllable)-1
			text.WriteString(utones[syllable[last:]])
			text.WriteString(syllable[1:last])
			continue
		}
		m := syllableRegexp.FindStringSubmatch(syllable)
		//if syllables == "m2 sha2" {fmt.Fprintf(os.Stderr, "  %#v\n", m)}
		if len(m)==0 {fmt.Fprintf(os.Stderr, "%#v: %#v\n", syllables, syllable)}
		t := tonify(m[2], m[4])
		text.WriteString(m[1])
		text.WriteString(t)
		text.WriteString(m[3])
	}
	return text.String()
}
