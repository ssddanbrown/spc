package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

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

type check struct {
	URL    string
	Check  string
	Passed bool
}

// TODO - Improve struct/prop names
// TODO - Show URL's that weren't checked

func main() {

	flag.Parse()
	args := flag.Args()

	def := loadDefinition(args)

	// Create check instances grouped by URL
	checkMap := make(map[string][]*check)
	checkCount := 0

	for _, checkDef := range def.Checks.CheckDefinitions {
		r, rErr := regexp.Compile(checkDef.URLRegex)
		if rErr != nil {
			errorAndExit(fmt.Sprintf("Error with check regex [%s]:\n%s", checkDef.URLRegex, rErr.Error()))
		}
		for _, url := range def.URLs {
			matches := r.FindStringSubmatch(url)
			if len(matches) > 0 {
				for _, checkStr := range checkDef.ChecksStrings {

					// Perform any applicable regex replaces in string
					for i, submatch := range matches {
						placeholder := fmt.Sprintf("$%d", i)
						checkStr = strings.Replace(checkStr, placeholder, submatch, -1)
					}

					c := check{
						URL:    url,
						Check:  checkStr,
						Passed: false,
					}
					checkMap[url] = append(checkMap[url], &c)
				}
				checkCount++
			}
		}

	}

	var wg sync.WaitGroup
	wg.Add(len(checkMap))

	for k, v := range checkMap {
		go checkSite(k, v, &wg)
	}

	fmt.Printf("\n\x1b[34mChecking %d urls, %d checks\x1b[0m\n\n", len(def.URLs), checkCount)

	wg.Wait()

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
		os.Exit(1)
	}

	fmt.Printf("\n\x1b[32m%s\x1b[0m\n", results)
	os.Exit(0)

}

func checkSite(url string, checks []*check, wg *sync.WaitGroup) {
	defer wg.Done()
	res, err := http.Get(url)
	if err != nil {
		// TODO - Add error message to check
		fmt.Println(err.Error())
		return
	}

	html, err := ioutil.ReadAll(res.Body)

	if err != nil {
		// TODO - Add error message to check
		fmt.Println(err.Error())
		return
	}

	for _, check := range checks {
		check.Passed = bytes.Contains(html, []byte(check.Check))
	}
}

func loadDefinition(args []string) definition {
	var err error
	var defContent []byte
	if len(args) == 0 {
		args = append(args, "")
	}

	path := args[0]

	if path == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			errorAndExit("No definition file provided")
		}
		defContent, err = ioutil.ReadAll(os.Stdin)
	} else {
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
