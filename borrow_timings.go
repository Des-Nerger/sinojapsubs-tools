package main

import (
	"fmt"
	"os"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}
func panicIfNotNil(e error) {if e!=nil {panic(e)}}

func main() {
	var (subs [2]*astisub.Subtitles; e error)
	for i:=0; i<2; i++ {
		subs[i], e = astisub.OpenFile(os.Args[1+i]); panicIfNotNil(e)
	}
	if len(subs[0].Items) != len(subs[1].Items) {
		panic(fmt.Sprintf("%v != %v", len(subs[0].Items), len(subs[1].Items)))
	}
	for i, it1 := range subs[1].Items {
		it0 := subs[0].Items[i]
		it1.StartAt, it1.EndAt = it0.StartAt, it0.EndAt
	}
	e = subs[1].Write(os.Args[1+2]); panicIfNotNil(e)
}
