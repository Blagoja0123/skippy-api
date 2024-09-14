package scraper

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Blagoja0123/skippy/models"
	"github.com/gocolly/colly"
)

type ForbesBSNScraper struct {
	domain     string
	categories []string
}

const ForbesImage = "https://seekvectors.com/files/download/Forbes-Logo-20.png"

func NewForbesBSNScraper(domain string, categories []string) *ForbesBSNScraper {
	return &ForbesBSNScraper{
		domain:     domain,
		categories: categories,
	}
}

func (fs *ForbesBSNScraper) Collect() []models.Article {

	results := make(chan []models.Article, len(fs.categories))

	var wg sync.WaitGroup

	articles := []models.Article{}
	timeStart := time.Now()
	for _, cat := range fs.categories {
		wg.Add(1)
		go func(category string) {
			defer wg.Done()
			collectForbesBSNCategory(cat, results)
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
	log.Printf("%d forbes articles scraped in %d ms", len(articles), timeEnd.UnixMilli()-timeStart.UnixMilli())
	return articles
}

func collectForbesBSNCategory(cat string, res chan []models.Article) {

	var articles []models.Article
	c := colly.NewCollector()

	c.OnHTML("div[data-test-e2e=\"stream articles\"]", func(e *colly.HTMLElement) {
		e.ForEach("div[data-testid=\"Card Stream\"]", func(_ int, el *colly.HTMLElement) {

			var tempArticle models.Article

			tempArticle.Title = el.ChildText("h3 > a")
			tempArticle.Origin = el.ChildAttr("h3 > a", "href")
			tempArticle.Content = el.ChildText("p")
			tempArticle.ImageURL = ForbesImage
			tempArticle.CategoryID = 4
			tempArticle.Source = "Forbes"
			date := strings.Split(el.ChildText("div > span"), "By")
			unixTime, _ := time.Parse("Jan 2, 2006", date[0])
			tempArticle.CreatedAt = unixTime
			articles = append(articles, tempArticle)
		})
	})

	c.Visit("https://www.forbes.com/" + cat + "/")

	res <- articles
}
