package main

/*
	usage examples:

	$ ffmpeg -y -v quiet -i "$B".??[!ast] -map 0:v -filter:v "crop=608:86:154:395" -f rawvideo -pix_fmt yuv444p - | go run extract_timings_from_hardsubs.go -fgYuvRanges d0 -frameSize 156864 -fps 30 "$B".timings.srt

	$ ffmpeg -y -v quiet -ss 00:00:00 -i "$B".??[!ast] -map 0:v -filter:v "crop=608:86:154:395" -f rawvideo -pix_fmt yuv444p - | go run extract_timings_from_hardsubs.go -fgYuvRanges d0 -frameSize 156864 -fps 30 /dev/null | avplay -loglevel quiet -f rawvideo -pixel_format yuv444p -video_size 608x86 - >/dev/null
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
	minFrameCount, maxFgPixelCountDiff, minFgPixelCount := 13, 400, 80
	var frameSize int; flag.IntVar(&frameSize, "frameSize", -1, "")
	var fps float64; flag.Float64Var(&fps, "fps", 0., "")
	r := []byte(nil)
	{
		var fgYuvRanges string; flag.StringVar(&fgYuvRanges, "fgYuvRanges", "", "")
		flag.Parse()
		for i, s := range strings.Split(fgYuvRanges, ",") {
			r = append(r, 0x00,0xFF, 0x00,0xFF, 0x00,0xFF)
			for j:=0; j<len(s); j+=2 {
				u, e := strconv.ParseUint(s[j:j+2], 16, 8); if e!=nil {continue}
				r[i*3*2 + j/2] = byte(u)
			}
		}
	}
	frameDelay := time.Duration(math.Round((1000*1000*1000) / fps))
	frameSizeThird := frameSize / 3; if frameSizeThird*3 != frameSize {panic(nil)}
	b,frameCount,fgPixelCount,posSum,head := make([]byte,frameSize),0,0,0,struct{frameCount,fgPixelCount,averagePos int}{}
	subs := astisub.NewSubtitles()
	handleHead := func() {
		if head.fgPixelCount>=minFgPixelCount && head.frameCount>=minFrameCount {
			subs.Items = append(subs.Items, &astisub.Item{
				StartAt: time.Duration(frameCount-head.frameCount)*frameDelay, EndAt: time.Duration(frameCount)*frameDelay,
				Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: fmt.Sprint(
					len(subs.Items)+1, head.frameCount, head.fgPixelCount)}}}},
			})
		}
	}
	for ;; frameCount++ {
		_, e := io.ReadFull(os.Stdin, b); if e==io.EOF {handleHead(); break}; panicIfNotNil(e)
		fgPixelCount,posSum=0,0
		for y,u,v:=0,frameSizeThird,2*frameSizeThird; y<frameSizeThird; y,u,v=y+1,u+1,v+1 {
			Y, U, V := b[y], b[u], b[v]
			/*b[y], b[u], b[v]*/ _, _, _ = func() (byte, byte, byte) {
				for j:=0; j<len(r); j+=3*2 {
					if r[j+0*2+0]<=Y && Y<=r[j+0*2+1] &&
					   r[j+1*2+0]<=U && U<=r[j+1*2+1] &&
					   r[j+2*2+0]<=V && V<=r[j+2*2+1] {
						fgPixelCount++; posSum+=y; return Y, U, V
					}
				}
				return 0x00, 0x80, 0x80
			} ()
		}
		averagePos := func() int {
			if fgPixelCount==0 {return 0}
			return posSum / fgPixelCount
		} ()
	//*
		if flag.Arg(0)=="/dev/null" && fgPixelCount >= 50 {
			fmt.Fprintln(os.Stderr, fgPixelCount, head.frameCount, head.fgPixelCount, abs(averagePos - head.averagePos))
			os.Stdout.Write(b)
		}
	/**/
		if abs(fgPixelCount-head.fgPixelCount) <= maxFgPixelCountDiff &&
		   abs(averagePos-head.averagePos) <= 1000 {
			head.frameCount++; continue
		}
		handleHead()
		head.fgPixelCount=fgPixelCount; head.averagePos=averagePos
		head.frameCount=1
	}
	for subs.Items[0].StartAt <= 3*time.Second {subs.Items = subs.Items[1:]}
	titlesAt := time.Duration(float64(frameCount) / fps * float64(time.Second)) - 30*time.Second
	for {
		lastIndex := len(subs.Items)-1
		if subs.Items[lastIndex].EndAt <= titlesAt {break}
		subs.Items = subs.Items[:lastIndex]
	}
	e := subs.Write(flag.Arg(0)); panicIfNotNil(e)
}
