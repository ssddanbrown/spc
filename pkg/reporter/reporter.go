package reporter

import (
	"fmt"
	"io"

	"github.com/ssddanbrown/spc/pkg/checker"
)

// Reporter is a type that is able to report on the checking results
type Reporter interface {
	Report(checker.CheckList, io.Writer)
}

// GetDefault provides the app default result Reporter
func GetDefault() Reporter {
	return terminalReporter{}
}

type terminalReporter struct{}

func (t terminalReporter) Report(checkMap checker.CheckList, w io.Writer) {
	// Print results
	var passes int
	var fails int

	for _, page := range checkMap {
		fmt.Fprintf(w, "\x1b[36m%s\x1b[0m\n", page.Path)
		for _, check := range page.Checks {
			countStr := fmt.Sprintf("%d", check.NeedleCount)
			if check.NeedleCount < 0 {
				countStr = "1+"
			}

			if check.Pass {
				fmt.Fprintf(w, "\t\x1b[32m✔ [%s] #%s\x1b[0m\n", check.Needle, countStr)
				passes++
			} else {
				fmt.Fprintf(w, "\t\x1b[31m✗ [%s] #%s\x1b[0m\n", check.Needle, countStr)
				fails++
			}
		}
	}

	results := fmt.Sprintf("%d checks passed, %d checks failed", passes, fails)
	passRate := (float32(passes) / float32(passes+fails)) * 100
	results += fmt.Sprintf(", %.2f%% of tests passed", passRate)

	if fails > 0 {
		fmt.Fprintf(w, "\n\x1b[31m%s\x1b[0m\n", results)
		return
	}

	fmt.Fprintf(w, "\n\x1b[32m%s\x1b[0m\n", results)
}
