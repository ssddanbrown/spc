package definition

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/ssddanbrown/spc/pkg/checker"
)

// Load the definition file into a map of checks
func Load(args []string) checker.CheckList {
	def := loadDefinition(args)

	pages := checker.CheckList{}

	for _, checkDef := range def.Checks.CheckDefinitions {
		r, rErr := regexp.Compile(checkDef.URLRegex)
		if rErr != nil {
			errorAndExit(fmt.Sprintf("Error with check regex [%s]:\n%s", checkDef.URLRegex, rErr.Error()))
		}
		for _, url := range def.URLs {
			page := checker.CheckedPage{Path: url}

			matches := r.FindStringSubmatch(url)
			if len(matches) == 0 {
				continue
			}

			for _, originalNeedle := range checkDef.ChecksStrings {
				needle := originalNeedle
				// Perform any applicable regex replaces in string
				for i, submatch := range matches {
					placeholder := fmt.Sprintf("$%d", i)
					needle = strings.Replace(needle, placeholder, submatch, -1)
				}

				c := &checker.Check{
					Needle:         needle,
					OriginalNeedle: originalNeedle,
					Pass:           false,
				}
				page.Checks = append(page.Checks, c)
			}

			pages = append(pages, page)
		}

	}
	return pages
}

type checkDefinitionList struct {
	CheckDefinitions []checkDefinition
}

func (l *checkDefinitionList) UnmarshalJSON(data []byte) error {
	var temp interface{}

	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	checkMap := temp.(map[string]interface{})

	for k, v := range checkMap {
		var cd checkDefinition
		cd.URLRegex = k

		// If just a string add to checks
		if cs, ok := v.(string); ok {
			cd.ChecksStrings = append(cd.ChecksStrings, cs)
		}

		// If an array of strings loop through and add the checks
		if css, ok := v.([]interface{}); ok {
			for _, s := range css {
				if cs, ok := s.(string); ok {
					cd.ChecksStrings = append(cd.ChecksStrings, cs)
				}
			}
		}

		if len(cd.ChecksStrings) > 0 {
			l.CheckDefinitions = append(l.CheckDefinitions, cd)
		}
	}

	return err
}

type checkDefinition struct {
	URLRegex      string
	ChecksStrings []string
}

type definition struct {
	Checks checkDefinitionList `json:"checks"`
	URLs   []string            `json:"urls"`
}

func loadDefinition(args []string) definition {
	var err error
	var defContent []byte
	if len(args) == 0 {
		args = append(args, "")
	}

	path := args[0]

	if path == "" {
		// Try loading in from StdIn
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			errorAndExit("No definition file provided")
		}
		defContent, err = ioutil.ReadAll(os.Stdin)
	} else {
		// Try loading in from file if path sensible length
		if len(path) < 300 {
			defContent, err = ioutil.ReadFile(path)
		} else {
			defContent = []byte(path)
		}
	}

	if err != nil {
		// Try using first argument as definition if json-looking
		if os.IsNotExist(err) && path[0] == '{' {
			defContent = []byte(path)
		} else {
			errorAndExit("Error when reading definition\n" + err.Error())
		}
	}

	var def definition
	err = json.Unmarshal(defContent, &def)
	if err != nil {
		errorAndExit(err.Error())
	}
	return def
}

func errorAndExit(m string) {
	fmt.Println(m)
	os.Exit(1)
}
