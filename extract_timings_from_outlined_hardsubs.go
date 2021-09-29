package main

/*
	usage examples:

	$ ffmpeg -y -v quiet -i "$B".??[!ast] -map 0:v -filter:v "crop=$W:$H:$X:$Y" -f rawvideo -pix_fmt gray - | go run extract_timings_from_outlined_hardsubs.go -w $W -h $H -fps $FPS "$B".srt

	$ ffmpeg -y -v quiet -ss 00:08:24 -i "$B".??[!ast] -map 0:v -filter:v "crop=$W:$H:$X:$Y" -f rawvideo -pix_fmt gray - | go run extract_timings_from_outlined_hardsubs.go -w $W -h $H -fps $FPS /dev/null | avplay -loglevel quiet -f rawvideo -pixel_format gray -video_size ${W}x$H - >/dev/null
*/

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/Des-Nerger/go-astisub"
	"github.com/asticode/go-astilog"
)

func init() {astilog.SetLogger(astilog.New(astilog.Configuration{Verbose: true}))}
func panicIfNotNil(e error) {if e!=nil {panic(e)}}
func abs(a int) int {m:=a>>(strconv.IntSize-1); return(a + m)^m}

type window struct {
	samples [2]int
	pos, sum int
}
func (w *window) updateSum(sample int) {
	w.sum -= w.samples[w.pos]
	w.samples[w.pos] = sample
	w.sum += w.samples[w.pos]
	w.pos++; if w.pos >= len(w.samples) {w.pos = 0}
}

func main() {
	minFrameCount, maxFgPixelCountDiff, minFgPixelCount := /*13*/30, 600, 80
	var frameWidth int; flag.IntVar(&frameWidth, "w", 0, "")
	var fps float64; flag.Float64Var(&fps, "fps", 0., "")
	frameSize := func() int {
		var frameHeight int; flag.IntVar(&frameHeight, "h", 0, "")
		flag.Parse()
		return frameWidth*frameHeight
	} ()
	frameDelay := time.Duration(math.Round((1000*1000*1000) / fps))
	b,frameCount,fgPixelCount,posSum,head := make([]byte,frameSize),0,0,0,struct{frameCount,fgPixelCount,averagePos int}{}
	subs := astisub.NewSubtitles()
	handleHead := func() {
		if head.fgPixelCount>=minFgPixelCount && head.frameCount>=minFrameCount {
			subs.Items = append(subs.Items, &astisub.Item{
				StartAt: time.Duration(frameCount-head.frameCount)*frameDelay, EndAt: time.Duration(frameCount)*frameDelay,
				Lines: []astisub.Line{{Items: []astisub.LineItem{{Text: fmt.Sprint(
					len(subs.Items)+1, head.frameCount, head.fgPixelCount, head.averagePos)}}}},
			})
		}
	}
	var w window
	for ;; frameCount++ {
		_, e := io.ReadFull(os.Stdin, b); if e==io.EOF {handleHead(); break}; panicIfNotNil(e)
		fgPixelCount,posSum=0,0
		var prevY byte
		for y:=0; y<frameSize; y++ {
			Y := b[y]
			if y % frameWidth == 0 {w, prevY = window{}, Y}
			w.updateSum(int(Y)-int(prevY))
			b[y]/*_*/ = func() byte {
				if abs(w.sum)>=160 {fgPixelCount++; posSum+=y; return Y}
				return 0x00
			} ()
			prevY = Y
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
		   abs(averagePos-head.averagePos) <= 1250 {
			head.frameCount++; continue
		}
		handleHead()
		head.fgPixelCount=fgPixelCount; head.averagePos=averagePos
		head.frameCount=1
	}
	e := subs.Write(flag.Arg(0)); panicIfNotNil(e)
}
