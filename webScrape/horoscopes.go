package webScrape

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
)

func ScrapeSign(sign string) (horoscope string) {
	c := colly.NewCollector()

	// On every p element which has style attribute call callback
	c.OnHTML("p[style]", func(e *colly.HTMLElement) {
		link := e.Attr("font-size:16px;")

		// Print link
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)

		if e.Text != "" {
			horoscope = e.Text
		}
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on https://www.ganeshaspeaks.com
	err := c.Visit("https://www.ganeshaspeaks.com/horoscopes/daily-horoscope/" + sign + "/")
	if err != nil {
		log.Fatal(err)
	}

	return horoscope
}
