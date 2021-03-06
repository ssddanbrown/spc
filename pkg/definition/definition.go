package definition

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ssddanbrown/spc/pkg/checker"
)

// Load the definition file into a map of checks
func Load(args []string) checker.CheckList {
	def, loadType := loadDefinition(args)
	def.Paths = append(def.Paths, def.URLs...)

	basePath, err := os.Getwd()
	if loadType == "path" {
		basePath = path.Dir(args[0])
	} else {
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	pages := checker.CheckList{}

	for _, pagePath := range def.Paths {
		isHTTP := (strings.Index(pagePath, "http://") == 0 || strings.Index(pagePath, "https://") == 0)

		if !isHTTP {
			fullPath := pagePath
			if fullPath[0] != '/' {
				fullPath = path.Join(basePath, fullPath)
			}

			files, err := filepath.Glob(fullPath)
			if err != nil {
				errorAndExit(fmt.Sprintf("Bad file path or glob pattern (%s)", fullPath))
			}
			for _, file := range files {
				page := parsePath(file, def.Checks.CheckDefinitions)
				pages = append(pages, page)
			}
			continue
		}

		page := parsePath(pagePath, def.Checks.CheckDefinitions)
		pages = append(pages, page)
	}

	return pages
}

func parsePath(path string, checkDefs []checkDefinition) checker.CheckedPage {
	page := checker.CheckedPage{Path: path}

	for _, checkDef := range checkDefs {
		r, rErr := regexp.Compile(checkDef.URLRegex)
		if rErr != nil {
			errorAndExit(fmt.Sprintf("Error with check regex [%s]:\n%s", checkDef.URLRegex, rErr.Error()))
		}

		matches := r.FindStringSubmatch(path)
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
	return page
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

		// If an array
		if arr, ok := v.([]interface{}); ok {
			for _, item := range arr {

				// Of check items
				if m, ok := item.(map[string]interface{}); ok {
					if ci, ok := createCheckItemFromMap(m); ok {
						cd.Checks = append(cd.Checks, ci)
					}
				}

				// Of strings
				if needle, ok := item.(string); ok {
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
	Paths  []string            `json:"paths"`
}

func loadDefinition(args []string) (definition, string) {
	var err error
	var defContent []byte
	var loadType string
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
		loadType = "stdin"
	} else {
		// Try loading in from file if path sensible length
		if len(path) < 300 {
			defContent, err = ioutil.ReadFile(path)
			loadType = "path"
		} else {
			defContent = []byte(path)
		}
	}

	if err != nil {
		// Try using first argument as definition if json-looking
		if os.IsNotExist(err) && path[0] == '{' {
			defContent = []byte(path)
			loadType = "inline"
		} else {
			errorAndExit("Error when reading definition\n" + err.Error())
		}
	}

	var def definition
	err = json.Unmarshal(defContent, &def)
	if err != nil {
		errorAndExit(err.Error())
	}
	return def, loadType
}

func errorAndExit(m string) {
	fmt.Println(m)
	os.Exit(1)
}
