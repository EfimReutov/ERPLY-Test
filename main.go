package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	baseURL = "https://www.aaa.com/autorepair"
	listURL = "https://www.aaa.com/autorepair/locations/%d?radius=%d&itemcategory=&napalocations="
)

var (
	location = 10018
	radius   = 500
)

type ProfileData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"telephone"`
}

type ProfList map[string]*ProfileData

func req(url string) (io.Reader, error) {
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

func main() {
	if err := aaaParser(location, radius); err != nil {
		log.Fatal(err)
	}
}

func aaaParser(location, radius int) error {
	//if err := linkMapperAAA(openTestHTML("testLink.html"), list); err != nil {
	//
	//list := make(ProfList)
	//f := openTestHTML("testInfo.html")
	//s := new(ProfileData)
	//if err := infoMapper(f, s); err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%+v", s)

	list, err := getLinksAAA(fmt.Sprintf(listURL, location, radius))
	if err != nil {
		return err
	}

	for link, info := range list {
		log.Printf("Proccesing link: %s", link)
		if err = getInfoAAA(link, info); err != nil {
			if err == errors.New("unexpected end of JSON input") {
				if err = getInfoAAA(link, info); err != nil {
					log.Printf("link %s get info error %s", link, err)
					continue
				}
			}
			log.Printf("link %s get info error %s", link, err)
		}
	}

	return csvWriter(list)
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
		"Company Name",
		"Phone",
		"E-mail",
	}

	err = writer.Write(headers)
	if err != nil {
		return err
	}

	for _, v := range list {
		err = writer.Write([]string{v.Name, v.Phone, v.Email})
		if err != nil {
			return err
		}
	}

	return nil
}
