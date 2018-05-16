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

	for _, url := range def.URLs {
		page := checker.CheckedPage{Path: url}

		for _, checkDef := range def.Checks.CheckDefinitions {
			r, rErr := regexp.Compile(checkDef.URLRegex)
			if rErr != nil {
				errorAndExit(fmt.Sprintf("Error with check regex [%s]:\n%s", checkDef.URLRegex, rErr.Error()))
			}

			matches := r.FindStringSubmatch(url)
			if len(matches) == 0 {
				continue
			}

			for _, ci := range checkDef.Checks {
				needle := ci.Check
				// Perform any applicable regex replaces in string
				for i, submatch := range matches {
					placeholder := fmt.Sprintf("$%d", i)
					needle = strings.Replace(needle, placeholder, submatch, -1)
				}

				c := &checker.Check{
					Needle:         needle,
					OriginalNeedle: ci.Check,
					Pass:           false,
					NeedleCount:    ci.Count,
				}
				page.Checks = append(page.Checks, c)
			}

		}

		pages = append(pages, page)
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
		if needle, ok := v.(string); ok {
			ci := checkItem{Check: needle, Count: -1}
			cd.Checks = append(cd.Checks, ci)
		}

		// If check item object
		if m, ok := v.(map[string]interface{}); ok {
			if ci, ok := createCheckItemFromMap(m); ok {
				cd.Checks = append(cd.Checks, ci)
			}
		}

		// If an array of strings loop through and add the checks
		if needles, ok := v.([]interface{}); ok {
			for _, s := range needles {
				if needle, ok := s.(string); ok {
					ci := checkItem{Check: needle, Count: -1}
					cd.Checks = append(cd.Checks, ci)
				}
			}
		}

		if len(cd.Checks) > 0 {
			l.CheckDefinitions = append(l.CheckDefinitions, cd)
		}
	}

	return err
}

func createCheckItemFromMap(m map[string]interface{}) (checkItem, bool) {
	var result checkItem
	ok := true

	if checkVal, ok := m["check"]; ok {
		if check, ok := checkVal.(string); ok {
			result.Check = check
		}
	}
	if !ok {
		return result, false
	}

	if countVal, ok := m["count"]; ok {
		if count, ok := countVal.(float64); ok {
			result.Count = int(count)
		}
	}

	return result, ok
}

type checkDefinition struct {
	URLRegex string
	Checks   []checkItem
}

type checkItem struct {
	Check string `json:"check"`
	Count int    `json:"count"`
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
