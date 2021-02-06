package main

/*
	usage examples:

	$ ffmpeg -y -v quiet -i "$B".??[!ast] -map 0:v -filter:v "crop=486:82:79:357" -f rawvideo -pix_fmt yuv444p - | go run "$T"/extract_timings_from_hardsubs.go -bgYuvRange __85739c78a4 -fgYuvRanges 9f__7d__7d -frameSize 119556 "$B".srt

	$ ffmpeg -y -v quiet -ss 00:35:12 -i "$B".??[!ast] -map 0:v -filter:v "crop=486:82:79:357" -f rawvideo -pix_fmt yuv444p - | go run "$T"/extract_timings_from_hardsubs.go -bgYuvRange __85739c78a4 -fgYuvRanges 9f__7d__7d -frameSize 119556 /dev/null | avplay -loglevel quiet -f rawvideo -pixel_format yuv444p -video_size 486x82 - >/dev/null
*/

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}
func panicIfNotNil(e error) {if e!=nil {panic(e)}}
func abs(a int) int {m:=a>>(strconv.IntSize-1); return(a + m)^m}

func main() {
	minFrameCount, maxFgPixelCountDiff, minFgPixelCount, fps := 25, 70, 80, 29.97003
	frameDelay := time.Duration(math.Round((1000*1000*1000) / fps))
	var frameSize int; flag.IntVar(&frameSize, "frameSize", -1, "")
	r := []byte(nil)
	{
		var bgYuvRange string; flag.StringVar(&bgYuvRange, "bgYuvRange", "", "")
		var fgYuvRanges string; flag.StringVar(&fgYuvRanges, "fgYuvRanges", "", "")
		flag.Parse()
		for i, s := range append([]string{bgYuvRange}, strings.Split(fgYuvRanges, ",")...) {
			r = append(r, 0x00,0xFF, 0x00,0xFF, 0x00,0xFF)
			for j:=0; j<len(s); j+=2 {
				u, e := strconv.ParseUint(s[j:j+2], 16, 8); if e!=nil {continue}
				r[i*3*2 + j/2] = byte(u)
			}
		}
	}
	frameSizeThird := frameSize / 3; if frameSizeThird*3 != frameSize {panic(nil)}
	minBgPixelCount := (frameSizeThird * 4) / 5
	b, frameCount, fgPixelCount, head := make([]byte, frameSize), 0, 0, struct{frameCount,fgPixelCount int} {}
	subs := astisub.NewSubtitles()
	handleHead := func(noMatterWhat bool) (ok bool) {
		if head.fgPixelCount>=minFgPixelCount && head.frameCount>=minFrameCount {
			if !noMatterWhat && fgPixelCount > head.fgPixelCount {
				head.fgPixelCount = fgPixelCount; head.frameCount++; return false }
			subs.Items = append(subs.Items, &astisub.Item{
				StartAt: time.Duration(frameCount-head.frameCount)*frameDelay, EndAt: time.Duration(frameCount)*frameDelay,
				Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: fmt.Sprint(
					len(subs.Items)+1, head.frameCount, head.fgPixelCount)}}}},
			})
		}
		return true
	}
	for ;; frameCount++ {
		_, e := io.ReadFull(os.Stdin, b); if e==io.EOF {handleHead(true); break}; panicIfNotNil(e)
		fgPixelCount=0; bgPixelCount:=0
		for y,u,v:=0,frameSizeThird,2*frameSizeThird; y<frameSizeThird; y,u,v=y+1,u+1,v+1 {
			Y, U, V := b[y], b[u], b[v]
			/*b[y], b[u], b[v]*/ _, _, _ = func() (byte, byte, byte) {
				for j:=0; j<len(r); j+=3*2 {
					if r[j+0*2+0]<=Y && Y<=r[j+0*2+1] &&
					   r[j+1*2+0]<=U && U<=r[j+1*2+1] &&
					   r[j+2*2+0]<=V && V<=r[j+2*2+1] {
						if j>=3*2 {fgPixelCount++; return Y, U, V}; bgPixelCount++
					}
				}
				return 0x00, 0x80, 0x80
			} ()
		}
		if bgPixelCount < minBgPixelCount {fgPixelCount = 0}
	//*
		if flag.Arg(0)=="/dev/null" && fgPixelCount >= 50 {
			fmt.Fprintln(os.Stderr, bgPixelCount, "/", minBgPixelCount, fgPixelCount, head.frameCount, head.fgPixelCount)
			os.Stdout.Write(b)
		}
	/**/
		if abs(fgPixelCount-head.fgPixelCount) <= maxFgPixelCountDiff {head.frameCount++; continue}
		if handleHead(false) {
			head.fgPixelCount = fgPixelCount
			head.frameCount = 1
		}
	}
	e := subs.Write(flag.Arg(0)); panicIfNotNil(e)
}
