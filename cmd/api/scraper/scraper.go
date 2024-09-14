package scraper

import "github.com/Blagoja0123/skippy/models"

type Scraper interface {
	Collect() []models.Article
}
