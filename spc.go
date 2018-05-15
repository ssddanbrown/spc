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

// TODO - Improve struct/prop names
// TODO - Show URL's that weren't checked

func main() {

	flag.Parse()
	args := flag.Args()

	checkList := definition.Load(args)
	fmt.Printf("\n\x1b[34mChecking %d urls, %d checks\x1b[0m\n\n", checkList.PageCount(), checkList.CheckCount())

	success := checker.Run(checkList)

	defaultReporter := reporter.GetDefault()
	defaultReporter.Report(checkList)

	if !success {
		os.Exit(1)
	}

	os.Exit(0)
}
