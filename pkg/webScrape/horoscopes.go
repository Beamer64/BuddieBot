package webScrape

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
)

func ScrapeSign(sign string) (string, error) {
	horoscope := ""
	found := false
	signNum := ""

	switch strings.ToLower(sign) {
	case "aries":
		signNum = "1"
	case "taurus":
		signNum = "2"
	case "gemini":
		signNum = "3"
	case "cancer":
		signNum = "4"
	case "leo":
		signNum = "5"
	case "virgo":
		signNum = "6"
	case "libra":
		signNum = "7"
	case "scorpio":
		signNum = "8"
	case "sagittarius":
		signNum = "9"
	case "capricorn":
		signNum = "10"
	case "aquarius":
		signNum = "11"
	case "pisces":
		signNum = "12"
	}

	c := colly.NewCollector()
	// On every p element which has style attribute call callback
	c.OnHTML(
		"p", func(e *colly.HTMLElement) {
			// link := e.Attr("font-size:16px;")

			if !found {
				if e.Text != "" {
					horoscope = e.Text
					found = true
				}
			}
		},
	)

	// Before making a request print "Visiting ..."
	c.OnRequest(
		func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		},
	)

	// StartServer scraping on https://www.horoscope.com
	err := c.Visit("https://www.horoscope.com/us/horoscopes/general/horoscope-general-daily-today.aspx?sign=" + signNum)
	if err != nil {
		return "", nil
	}

	return horoscope, nil
}
