// update creature raws from en wiki
package main

import (
	"os"
	"strings"
	"sync"

	"cgt.name/pkg/go-mwclient"        // imports mwclient
	"cgt.name/pkg/go-mwclient/params" // imports params
)

func getPages(wiki mwclient.Client, categoryName string) []string {
	var rawList []string
	req := params.Values{
		"list":    "categorymembers",
		"cmtitle": categoryName,
	}
	queryCategory := wiki.NewQuery(req)

	if queryCategory.Err() != nil {
		panic(queryCategory.Err())
	}

	for queryCategory.Next() {
		object, err := queryCategory.Resp().GetObject("query")
		if err != nil {
			panic(err)
		}

		objectArray, err := object.GetObjectArray("categorymembers")
		if err != nil {
			panic(err)
		}

		for _, val := range objectArray {
			title, err := val.GetString("title")
			if err != nil {
				panic(err)
			}
			rawList = append(rawList, title)
		}
	}
	
	return rawList
}

func updateRaw(wg *sync.WaitGroup, wikiRu mwclient.Client, wikiEn mwclient.Client, title string) {
	defer wg.Done()
	content, _, err := wikiEn.GetPageByName(title)
	if err != nil {
		println(title, " get failed: ", err.Error())
		return
	}

	content = strings.ReplaceAll(content, "v50:", ":")
	req := params.Values{
		"title":   title,
		"text":    content,
		"summary": "raw update",
	}
	err = wikiRu.Edit(req)

	if err != nil && err.Error() != "edit successful, but did not change page" {
		println(title, " update failed: ", err.Error())
		return
	}

}
func main() {
	if len(os.Args) != 3 {
		panic("usage: " + os.Args[0] + " dfwk_login password")
	}

	wikiRu, err := mwclient.New("https://dfwk.ru/api.php", "dfwk.ru-MEX_BOT")
	if err != nil {
		panic(err)
	}

	wikiEn, err := mwclient.New("https://dwarffortresswiki.org/api.php", "dfwk.ru-MEX_BOT")
	if err != nil {
		panic(err)
	}

	err = wikiRu.Login(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}

	ruRawList := getPages(*wikiRu, "Категория:Raw-файлы_существ")

	var wg sync.WaitGroup
	var threadCount = 10

	for i := range ruRawList {
		title := ruRawList[i]
		wg.Add(1)
		go updateRaw(&wg, *wikiRu, *wikiEn, title)
		if i%threadCount == 0 { // limit concurrent goroutines since wiki can bug out and start to refuse edits
			wg.Wait()
		}
	}
	wg.Wait()
}
