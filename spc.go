package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ssddanbrown/spc/pkg/checker"
	"github.com/ssddanbrown/spc/pkg/definition"
	"github.com/ssddanbrown/spc/pkg/reporter"
)

// type Results struct {
// 	Pages []CheckedPage
// 	Pass  bool
// }

// type CheckedPage struct {
// 	URI     string
// 	Content string
// 	Pass    bool
// }

// type Check struct {
// 	Needle         string
// 	NeedleCount    int
// 	OriginalNeedle string
// 	Pass           bool
// }

// TODO - Improve struct/prop names
// TODO - Show URL's that weren't checked

func main() {

	flag.Parse()
	args := flag.Args()

	checkMap := definition.Load(args)

	// fmt.Printf("\n\x1b[34mChecking %d urls, %d checks\x1b[0m\n\n", len(def.URLs), checkCount)

	checker.Run(checkMap)

	defaultReporter := reporter.GetDefault()
	success := defaultReporter.Report(checkMap)

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}

func errorAndExit(m string) {
	fmt.Println(m)
	os.Exit(1)
}
