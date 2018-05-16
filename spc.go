package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ssddanbrown/spc/pkg/checker"
	"github.com/ssddanbrown/spc/pkg/definition"
	"github.com/ssddanbrown/spc/pkg/reporter"
)

// TODO - Add tests
// TODO - Add help and version command
// TODO - Show URL's that weren't checked

func main() {

	flag.Parse()

	// Load definition
	args := flag.Args()
	checkList := definition.Load(args)

	// Run checks
	fmt.Printf("\n\x1b[34mChecking %d urls, %d checks\x1b[0m\n\n", checkList.PageCount(), checkList.CheckCount())
	success := checker.Run(checkList)

	// Report on the results
	defaultReporter := reporter.GetDefault()
	defaultReporter.Report(checkList, os.Stdout)

	if !success {
		os.Exit(1)
	}
	os.Exit(0)
}
