package rubytagpolyfill

import (
	"encoding/xml"
	"fmt"
	"io"
	//"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	. "unsafe"
)

var (
	r *strings.Reader
	d *xml.Decoder
	resetD func()
)

func init() {
	r = &strings.Reader{}
	d = xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	{
		d := reflect.ValueOf(d).Elem()
		dErr, dOffset :=
			(*error)(Pointer(d.FieldByName("err").UnsafeAddr())),
			(*int64)(Pointer(d.FieldByName("offset").UnsafeAddr()))
		resetD = func() {*dErr, *dOffset = nil, 0}
	}
}

const (
	space = " "//"-"//
	spaceSize = len(space)
	maxRuneWidthSpaces = space + space + space + space
)
func widthRune(r rune) (i int) {
	//defer func() {fmt.Printf("%c %v\n", r, i)} ()
	if r < '\u0800' {i=1; return}
	if unicode.Is(unicode.Han, r) {i=len(maxRuneWidthSpaces)/spaceSize; return}
	i=2; return
}
func widthString(s string) (width int) {
	for _, r := range s {width += widthRune(r)}
	return
}

func signBitAndAbs(n int) (signBit, _ int) {
	signBit = n >> (strconv.IntSize-1)
	return signBit, (n ^ signBit) - signBit
}

func centerAlign(high, low string) (newHigh, newLow string) {
	highWidth, lowWidth := widthString(high), widthString(low)
	newHigh, newLow = high, low
	if highWidth==lowWidth {return}
	signBit, absDiff := signBitAndAbs(lowWidth - highWidth)
	halfAbsDiff := absDiff/2
	padding := strings.Repeat(space, absDiff - halfAbsDiff)
	if signBit==0 {
		newHigh = padding[:halfAbsDiff*spaceSize] + high + padding
		return
	}
	newLow = padding[:halfAbsDiff*spaceSize] + low + padding
	return
}

func ToSpacefilledTwoLines(line, doubleSizeBegin, doubleSizeEnd string) []string {
	if !strings.Contains(line, "<ruby>") {return []string{line}}
	r.Reset(line)
	resetD()
	inRuby, inRt, inRp := false, false, false
	sb := strings.Builder{}; sb.Grow(len(line)*3/2)
	ses := []string(nil)
loop:
	for {
		inputOffset := d.InputOffset()
		tok, err := d.RawToken()
		switch err {
		case nil:
		case io.EOF: break loop
		default: panic(err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			switch tok.Name.Local {
			case "ruby": inRuby=true
			case "rt": inRt=true
			case "rb": inRt=false
			case "rp": inRp=true
			}
		case xml.EndElement:
			switch tok.Name.Local {
			case "ruby": inRuby=false; fallthrough
			case "rt": inRt=false
			case "rp": inRp=false
			}
		case xml.CharData:
			if inRp {break}
			s := line[inputOffset:d.InputOffset()]
			if inRt {
				last := len(ses)-1
				s, newSesLast := centerAlign(s, ses[last])
				i := strings.Index(newSesLast, ses[last])
				ses[last] = newSesLast[:i] + doubleSizeBegin + ses[last] + doubleSizeEnd + newSesLast[i+len(ses[last]):]
				sb.WriteString(s)
			} else {
				ses = append(ses, s)
				if !inRuby {
					for s := s; s!=""; {
						r, size := utf8.DecodeRuneInString(s)
						if size==0 {panic(fmt.Sprintf("error while rune decoding %#v", s))}
						sb.WriteString(maxRuneWidthSpaces[:widthRune(r)*spaceSize])
						s = s[size:]
					}
				}
			}
		}
	}
	return []string{sb.String(), strings.Join(ses, "")}
}
