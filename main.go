package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const captchaTime = time.Second * 7

type ProfileData struct {
	CompanyName      string
	LegalForm        string
	RegistryCode     string
	RegistrationDate string
	PreviousNames    string
	FieldOfOperation string
	Capital          string
	Address          string
	Status           string
	Email            string
}

type ProfList map[string]*ProfileData

func main() {
	list, n, err := start()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Total pages: ", n)
	n = 1

	if err = cont(list, n); err != nil {
		log.Fatalln(err)
	}

	for link, info := range list {
		if err = infoReq(link, info); err != nil {
			log.Fatalln(err)
		}
	}

	if err = csvWriter(list); err != nil {
		log.Fatalln(err)
	}
}

func start() (ProfList, int, error) {
	body, err := linksReq(0)
	if err != nil {
		return nil, 0, err
	}

	buf := new(bytes.Buffer)
	n, err := getNumberOfPages(io.TeeReader(body, buf))
	if err != nil {
		return nil, 0, err
	}

	list := make(ProfList, n*10)

	if err = linkMapper(buf, list); err != nil {
		return nil, 0, err
	}

	return list, n, nil
}

func cont(list ProfList, n int) error {
	var body io.Reader
	var err error
	for i := 2; i <= n; i++ {
		if body, err = linksReq(i); err != nil {
			return err
		}
		if err = linkMapper(body, list); err != nil {
			return err
		}
	}
	return nil
}

func linksReq(page int) (io.Reader, error) {
	p := ""
	if page > 1 {
		p = fmt.Sprintf("/%d", page)
	}
	return req("https://www.teatmik.ee/en/advancedsearch/business/eyJhcyI6WyI0NSJdfQ==" + p)
}

func infoReq(url string, s *ProfileData) error {
	body, err := req(url)
	if err != nil {
		return err
	}
	return infoMapper(body, s)
}

func req(url string) (io.Reader, error) {
	time.Sleep(captchaTime)

	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func getNumberOfPages(body io.Reader) (int, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return 0, err
	}

	f := func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("onchange")
		return strings.HasPrefix(res, "window")
	}

	tag := doc.Find("body select").FilterFunction(f).Find("option").Last()

	return strconv.Atoi(tag.Text())
}

func linkMapper(body io.Reader, list ProfList) error {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	f := func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("class")
		return strings.EqualFold(res, "search-res")
	}

	doc.Find("body a").FilterFunction(f).Each(func(_ int, tag *goquery.Selection) {
		link, _ := tag.Attr("href")
		list[link] = &ProfileData{}
	})
	return nil
}

func infoMapper(body io.Reader, s *ProfileData) error {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}

	s.CompanyName = doc.Find("body h1").Text()

	f := func(i int, s *goquery.Selection) bool {
		res, _ := s.Attr("class")
		return strings.EqualFold(res, "info")
	}

	doc.Find("body table").FilterFunction(f).Find("tr").Each(func(_ int, tag *goquery.Selection) {
		row := tag.Find("td")
		switch row.First().Text() {
		case " Legal form:":
			s.LegalForm = row.Last().Text()
		case " Registry code:":
			s.RegistryCode = row.Last().Text()
		case " Registration date:":
			s.RegistrationDate = row.Last().Text()
		case " Previous names:":
			s.PreviousNames = row.Last().Text()
		case " Field of operation:":
			s.FieldOfOperation = row.Last().Text()
		case " Capital:":
			s.Capital = row.Last().Text()
		case " Address:":
			s.Address = row.Last().Text()
		case " Status:":
			s.Status = row.Last().Text()
		case " E-mail":
			s.Email = row.Last().Text()
		}
	})

	return nil
}

func csvWriter(list ProfList) error {
	file, err := os.Create("result.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"companyName",
		"Legal form",
		"Registry code",
		"Registration date",
		"Previous names",
		"Field of operation",
		"Capital",
		"Address",
		"Status",
		"E-mail",
	}

	err = writer.Write(headers)
	if err != nil {
		return err
	}

	for _, v := range list {
		err = writer.Write([]string{v.CompanyName, v.LegalForm, v.RegistryCode, v.RegistrationDate, v.PreviousNames, v.FieldOfOperation, v.Capital, v.Address, v.Status, v.Email})
		if err != nil {
			return err
		}
	}

	return nil
}
