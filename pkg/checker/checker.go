package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type Check struct {
	URL    string
	Check  string
	Passed bool
}

func Run(checkMap map[string][]*Check) {
	var wg sync.WaitGroup
	wg.Add(len(checkMap))

	for k, v := range checkMap {
		go checkSite(k, v, &wg)
	}

	wg.Wait()
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
		check.Passed = bytes.Contains(html, []byte(check.Check))
	}
}
