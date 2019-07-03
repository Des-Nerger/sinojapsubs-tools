package main

import (
	"unicode"
)

var (
	AsciiAlphaNum = &unicode.RangeTable{R16: []unicode.Range16{
		{'0', '9', 1},
		{'A', 'Z', 1},
		{'a', 'z', 1},
	}}
	FullwidthAlphaNum = &unicode.RangeTable{R16: []unicode.Range16{
		{'０', '９', 1},
		{'Ａ', 'Ｚ', 1},
		{'ａ', 'ｚ', 1},
	}}
	AlphaNum = &unicode.RangeTable{R16: append(AsciiAlphaNum.R16, FullwidthAlphaNum.R16...)}
)

func InitLatinOffsets(rangeTables ...*unicode.RangeTable) {
	for _, rangeTable := range rangeTables {
		for _, range16 := range rangeTable.R16 {
			if range16.Hi > unicode.MaxLatin1 {
				break
			}
			rangeTable.LatinOffset++
		}
	}
}

func init() {
	InitLatinOffsets(AsciiAlphaNum, FullwidthAlphaNum, AlphaNum)
}
