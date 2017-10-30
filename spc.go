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
	"sync"
)

type definition struct {
	Checks map[string]string `json:"checks"`
	URLs   []string          `json:"urls"`
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

	if len(args) < 1 {
		errorAndExit("No definition file provided")
	}

	def := loadDefinition(args[0])

	// Create check instances grouped by URL
	checkMap := make(map[string][]*check)
	checkCount := 0
	for checkRegex, checkStr := range def.Checks {
		r, rErr := regexp.Compile(checkRegex)
		if rErr != nil {
			errorAndExit(fmt.Sprintf("Error with check regex [%s]:\n%s", checkRegex, rErr.Error()))
		}
		for _, url := range def.URLs {
			if r.Match([]byte(url)) {
				c := check{
					URL:    url,
					Check:  checkStr,
					Passed: false,
				}
				checkMap[url] = append(checkMap[url], &c)
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

func loadDefinition(path string) definition {
	var err error
	var defContent []byte

	if len(path) < 300 {
		defContent, err = ioutil.ReadFile(path)
	} else {
		defContent = []byte(path)
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
