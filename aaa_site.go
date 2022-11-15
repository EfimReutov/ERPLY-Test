package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io"
	"strings"
)

func getLinksAAA(url string) (ProfList, error) {
	body, err := req(url)
	if err != nil {
		return nil, err
	}

	list := make(ProfList)

	if err = linkMapperAAA(body, list); err != nil {
		return nil, err
	}

	return list, nil
}

func linkMapperAAA(body io.Reader, list ProfList) error {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return err
	}

	f := func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("class")
		return strings.EqualFold(res, "aar-detail")
	}

	doc.Find("a").FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {
		link, _ := tag.Attr("href")
		list[strings.Replace(link, "..", baseURL, 1)] = &ProfileData{}
	})

	return nil
}

func getInfoAAA(url string, s *ProfileData) error {
	body, err := req(url)
	if err != nil {
		return err
	}
	return infoMapper(body, s)
}

func infoMapper(body io.Reader, s *ProfileData) error {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return err
	}

	f := func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("type")
		return strings.EqualFold(res, "application/ld+json")
	}

	sel := doc.Find("script").FilterFunction(f)

	return json.Unmarshal([]byte(sel.Text()), s)
}
