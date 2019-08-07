// Copyright 2019 Jacques Supcik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This package fetches the lake level information from the web site of
// the Groupe-e. There is no API, therfore we use scraping.

package gelakelevel

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const (
	pageURL       = "https://www.groupe-e.ch/fr/univers-groupe-e/niveau-lacs"
	gruyereURL    = "https://www.groupe-e.ch/fr/univers-groupe-e/niveau-lacs/gruyere"
	schiffenenURL = "https://www.groupe-e.ch/fr/univers-groupe-e/niveau-lacs/schiffenen"
)

type Measure struct {
	Date time.Time
	Min  float64
	Max  float64
}

type Lake struct {
	Name     string
	MaxLevel float64
	Measures map[string]Measure
}

type Lakes map[string]Lake

// msm parses a string representing a lake's level ang returns a float
// note: "msm" means "mètres sur mer" (in french) which means "metres above sea level"
func msm(t string) float64 {
	re := regexp.MustCompile(`(\d+\.\d+).*msm`)
	n := re.FindStringSubmatch(t)
	if n != nil {
		nf, err := strconv.ParseFloat(n[1], 64)
		if err == nil {
			return nf
		}
	}
	return 0
}

func getDetail(client *http.Client, url string, mes map[string]Measure) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	table := doc.Find("table").First()
	dates := table.Find("table").First().Find("td").Map(func(i int, s *goquery.Selection) string { return s.Text() })
	mins := table.Find("table").Eq(1).Find("td").Map(func(i int, s *goquery.Selection) string { return s.Text() })
	maxs := table.Find("table").Eq(2).Find("td").Map(func(i int, s *goquery.Selection) string { return s.Text() })

	if len(dates) != len(mins) || len(dates) != len(maxs) {
		return errors.New("array size mismatch")
	}
	for i := 0; i < len(dates); i++ {
		d, err := time.Parse("2.1.2006", strings.TrimSpace(dates[i]))
		if err != nil {
			return err
		}
		mes[d.Format("2006-01-02")] = Measure{d, msm(mins[i]), msm(maxs[i])}
	}
	return nil
}

// GetLevel returns the level of all lakes
func GetLevel(client *http.Client) (Lakes, error) {
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	result := make(Lakes)
	table := doc.Find("table").First()
	header := table.Find("thead tr th")
	date0, err := time.Parse("2.1.2006", strings.TrimSpace(header.Eq(2).Text()))
	if err != nil {
		return nil, err
	}
	date1, err := time.Parse("2.1.2006", strings.TrimSpace(header.Eq(3).Text()))
	if err != nil {
		return nil, err
	}
	body := table.Find("tbody tr")
	for i := 0; i < body.Length(); i++ {
		item := body.Eq(i)
		name := strings.TrimSpace(item.Find("td").Eq(0).Text())
		name = regexp.MustCompile("[*]").ReplaceAllString(name, "")
		maxLevel := msm(strings.TrimSpace(item.Find("td").Eq(1).Text()))
		l0 := msm(strings.TrimSpace(item.Find("td").Eq(2).Text()))
		l1 := msm(strings.TrimSpace(item.Find("td").Eq(3).Text()))
		lake := Lake{
			Name:     name,
			MaxLevel: maxLevel,
		}
		lake.Measures = make(map[string]Measure)
		lake.Measures[date0.Format("2006-01-02")] = Measure{date0, l0, l0}
		lake.Measures[date1.Format("2006-01-02")] = Measure{date1, l1, l1}
		if name == "La Gruyère" {
			err := getDetail(client, gruyereURL, lake.Measures)
			if err != nil {
				return nil, err
			}
		} else if name == "Schiffenen" {
			err := getDetail(client, schiffenenURL, lake.Measures)
			if err != nil {
				return nil, err
			}
		}
		result[name] = lake
	}
	return result, nil
}
