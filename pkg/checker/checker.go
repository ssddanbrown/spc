package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type CheckedPage struct {
	Path    string
	Content string
	Checks  []*Check
}

type Check struct {
	Needle         string
	NeedleCount    int
	OriginalNeedle string
	Pass           bool
}

// Run all tests against all pages
func Run(list CheckList) {
	var wg sync.WaitGroup
	wg.Add(len(list))

	for _, page := range list {
		go checkSite(page.Path, page.Checks, &wg)
	}

	wg.Wait()
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
		check.Pass = bytes.Contains(html, []byte(check.Needle))
	}
}
