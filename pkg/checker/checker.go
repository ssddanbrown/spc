package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"sync"
)

type CheckedPage struct {
	Path    string
	Content string
	Checks  []*Check
	Pass    bool
}

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
		go checkSite(page.Path, page.Checks, &wg)
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

type CheckList []CheckedPage

func (cl CheckList) PageCount() int {
	return len(cl)
}

func (cl CheckList) CheckCount() int {
	total := 0
	for _, list := range cl {
		total += len(list.Checks)
	}
	return total
}

func checkSite(url string, checks []*Check, wg *sync.WaitGroup) {
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
		containCount := bytes.Count(html, []byte(check.Needle))
		if check.NeedleCount < 0 {
			check.Pass = containCount > 0
			continue
		}

		check.Pass = containCount == check.NeedleCount
	}

}
