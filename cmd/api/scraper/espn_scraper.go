package scraper

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Blagoja0123/skippy/models"
	"github.com/gocolly/colly"
)

type ESPNScraper struct {
	domain     string
	categories []string
	// catService *service.CategoryService
	// artService *service.ArticleService
}

func NewESPNScraper(domain string, categories []string) *ESPNScraper {
	return &ESPNScraper{
		domain:     domain,
		categories: categories,
		// catService: cat,
		// artService: art,
	}
}

func (espn *ESPNScraper) Collect() []models.Article {

	results := make(chan []models.Article, len(espn.categories))

	var wg sync.WaitGroup

	articles := []models.Article{}
	timeStart := time.Now()
	for _, cat := range espn.categories {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()
			collectCategory(cat, results)
		}(cat)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for articleGroup := range results {
		articles = append(articles, articleGroup...)
	}
	timeEnd := time.Now()
	// log.Print(articles)
	log.Printf("%d espn articles scraped in %d ms", len(articles), timeEnd.UnixMilli()-timeStart.UnixMilli())
	return articles
}

func collectCategory(cat string, res chan []models.Article) {

	var articles []models.Article
	c := colly.NewCollector()

	c.OnHTML("section#news-feed", func(e *colly.HTMLElement) {
		e.ForEach("section.contentItem", func(_ int, el *colly.HTMLElement) {
			// Extract the title
			var tempArticle models.Article
			tempArticle.Title = el.ChildText("h2.contentItem__title")
			if tempArticle.Title == "" {
				return
			}
			tempArticle.Content = el.ChildText("p.contentItem__subhead")
			tempArticle.ImageURL = "https://static.vecteezy.com/system/resources/thumbnails/027/127/453/small_2x/espn-logo-espn-icon-transparent-free-png.png"
			tempArticle.Source = "ESPN"
			tempArticle.Origin = el.ChildAttr("a", "href")
			tempArticle.CategoryID = 5
			timestamp := el.ChildText("span.contentMeta__timestamp")
			if len(timestamp) > 3 {
				if strings.Contains(timestamp, "h") {
					timestamp = strings.Split(timestamp, "h")[0] + "h"
				} else {
					timestamp = strings.Split(timestamp, "d")[0] + "d"
				}
			}
			var artTime time.Time
			if timestamp[len(timestamp)-1] == 'h' {
				hour, _ := strconv.Atoi(timestamp[0 : len(timestamp)-1])

				currentTime := time.Now()
				artTime = currentTime.Add(time.Duration(-hour) * time.Hour)
			} else if timestamp[len(timestamp)-1] == 'd' {
				day, _ := strconv.Atoi(timestamp[0 : len(timestamp)-1])
				currentTime := time.Now()
				artTime = currentTime.Add(time.Duration(-day*24) * time.Hour)
			} else {
				artTime = time.Now()
			}
			tempArticle.CreatedAt = artTime

			if len(strings.Split(tempArticle.Origin, "/")) <= 3 {
				return
			}
			if tempArticle.Origin[0:4] == "http" {
				splitStrs := strings.Split(tempArticle.Origin, "http://www.espn.co.uk")
				if len(splitStrs) >= 2 {
					tempArticle.Origin = "https://www.espn.co.uk" + splitStrs[1]
				}
			} else {
				tempArticle.Origin = "https://www.espn.co.uk" + tempArticle.Origin
			}
			articles = append(articles, tempArticle)
		})
	})

	c.Visit("https://www.espn.com/" + cat)

	res <- articles
}
