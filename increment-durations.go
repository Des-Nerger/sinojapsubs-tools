package main
import (
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/Des-Nerger/sinojapsubs-tools/ensure_go_duration_format"
)

func main() {
	var (args=os.Args[1:]; increment time.Duration)
	for i, a := range args {
		if i!=0 && i%2==0 {continue}
		d, err := time.ParseDuration(EnsureGoDurationFormat(a))
		if err!=nil {panic(err)}
		if i==0 {
			increment = d
		} else {
			args[i] = (d+increment).String()
		}
	}
	fmt.Println(strings.Join(args[1:], " "))
}
