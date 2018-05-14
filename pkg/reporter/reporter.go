package reporter

import (
	"fmt"

	"github.com/ssddanbrown/spc/pkg/checker"
)

// Reporter is a type that is able to report on the checking results
type Reporter interface {
	Report(map[string][]*checker.Check) bool
}

// GetDefault provides the app default result Reporter
func GetDefault() Reporter {
	return terminalReporter{}
}

type terminalReporter struct{}

func (t terminalReporter) Report(checkMap map[string][]*checker.Check) bool {
	// Print results
	var passes int
	var fails int

	for url, checks := range checkMap {
		fmt.Printf("\x1b[36m%s\x1b[0m\n", url)
		for _, check := range checks {
			if check.Passed {
				fmt.Printf("\t\x1b[32m✔ [%s]\x1b[0m\n", check.Check)
				passes++
			} else {
				fmt.Printf("\t\x1b[31m✗ [%s]\x1b[0m\n", check.Check)
				fails++
			}
		}
	}

	results := fmt.Sprintf("%d checks passed, %d checks failed", passes, fails)
	passRate := (float32(passes) / float32(passes+fails)) * 100
	results += fmt.Sprintf(", %.2f%% of tests passed", passRate)

	if fails > 0 {
		fmt.Printf("\n\x1b[31m%s\x1b[0m\n", results)
		return false
	}

	fmt.Printf("\n\x1b[32m%s\x1b[0m\n", results)
	return true
}
