package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// CheckedPage contains information about a single page that was checked
type CheckedPage struct {
	Path    string
	Content string
	Checks  []*Check
	Pass    bool
}

// Check holds the details of a check
type Check struct {
	Needle         string
	NeedleCount    int
	OriginalNeedle string
	Pass           bool
}

// Run all tests against all pages
func Run(list CheckList) bool {
	var wg sync.WaitGroup
	wg.Add(len(list))

	for _, page := range list {
		go checkPage(page.Path, page.Checks, &wg)
	}

	wg.Wait()

	// Get overall status
	overallPass := true

	for _, page := range list {
		page.Pass = true
		for _, check := range page.Checks {
			if !check.Pass {
				page.Pass = false
				overallPass = false
			}
		}
	}

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Path < list[j].Path
	})

	return overallPass
}

// CheckList is a slice of CheckedPages
type CheckList []CheckedPage

// PageCount provides a count of the total number of pages checked
func (cl CheckList) PageCount() int {
	return len(cl)
}

// CheckCount provides a count of the total number of checks made
func (cl CheckList) CheckCount() int {
	total := 0
	for _, list := range cl {
		total += len(list.Checks)
	}
	return total
}

func checkPage(loc string, checks []*Check, wg *sync.WaitGroup) {
	defer wg.Done()

	var content []byte
	var err error
	if strings.Index(loc, "http://") == 0 || strings.Index(loc, "https://") == 0 {
		content, err = getWebpage(loc)
	} else {
		content, err = ioutil.ReadFile(loc)
	}

	if err != nil {
		// TODO - Add error message to check
		fmt.Println(err.Error())
		return
	}

	for _, check := range checks {
		containCount := bytes.Count(content, []byte(check.Needle))
		if check.NeedleCount < 0 {
			check.Pass = containCount > 0
			continue
		}

		check.Pass = containCount == check.NeedleCount
	}

}

func getWebpage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return ioutil.ReadAll(res.Body)
}
