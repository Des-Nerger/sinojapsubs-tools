package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	j := Jumanpp{Path: "/opt/jumanpp/bin/jumanpp"}
	j.Start()
	defer j.Wait()
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		fmt.Fprintf(os.Stderr, "%q\n", j.AnalyzeLine(sc.Text()))
	}
}
