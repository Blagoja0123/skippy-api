package provider

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
)

type NYTimesProvider struct {
	endpoint    string
	sections    []string
	queryParams map[string]string
	api_key     string
	catService  *service.CategoryService
	artService  *service.ArticleService
}

func NewNYTimesProvider(endpoint string, sections []string, queryParams map[string]string, api_key string, cat *service.CategoryService, art *service.ArticleService) *NYTimesProvider {
	return &NYTimesProvider{
		endpoint:    endpoint,
		sections:    sections,
		queryParams: queryParams,
		api_key:     api_key,
		catService:  cat,
		artService:  art,
	}
}

type Multimedia struct {
	URL     string `json:"url"`
	Format  string `json:"format"`
	Height  int    `json:"height"`
	Width   int    `json:"width"`
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Caption string `json:"caption"`
}

type NYTArticle struct {
	Title         string       `json:"title"`
	Abstract      string       `json:"abstract"`
	PublishedDate string       `json:"published_date"`
	URL           string       `json:"url"`
	Multimedia    []Multimedia `json:"multimedia"`
}

// Response represents the entire API response.
type Response struct {
	Status      string       `json:"status"`
	Copyright   string       `json:"copyright"`
	Section     string       `json:"section"`
	LastUpdated string       `json:"last_updated"`
	NumResults  int          `json:"num_results"`
	Results     []NYTArticle `json:"results"`
}

// https://api.nytimes.com/svc/search/v2/articlesearch.json?&api-key=7VLGGiPTqbfq7VFVHhQpudMYvwLGNhqW

func (nyt *NYTimesProvider) GetArticles(ctx context.Context) (map[string]interface{}, error) {

	data := make(map[string]interface{})

	var dbwrite int64 = 0
	var total uint = 0

	for _, section := range nyt.sections {

		res, err := http.Get(nyt.endpoint + section + ".json?api-key=" + nyt.api_key)
		if err != nil {
			return nil, err
		}
		body, _ := io.ReadAll(res.Body)
		var jsonRes Response

		if err := json.Unmarshal(body, &jsonRes); err != nil {
			return nil, err
		}
		for _, doc := range jsonRes.Results {

			total++

			var category *models.Category
			var err error

			if section == "sports" {
				category, err = nyt.catService.GetByName(ctx, "sport")
			} else if section == "movies" {
				category, err = nyt.catService.GetByName(ctx, "film")
			} else {
				category, err = nyt.catService.GetByName(ctx, section)
			}

			if err != nil {
				return nil, err
			}

			pubDate, _ := time.Parse(time.UnixDate, doc.PublishedDate)
			var imageURL string

			for _, media := range doc.Multimedia {
				if media.Format == "Large Thumbnail" {
					imageURL = media.URL
					break
				}
			}

			newArticle := models.Article{
				Title:      doc.Title,
				Content:    doc.Abstract,
				Source:     "The New York Times",
				Origin:     doc.URL,
				CreatedAt:  pubDate,
				ImageURL:   imageURL,
				CategoryID: category.ID,
			}

			start := time.Now().UnixMilli()
			err = nyt.artService.Create(ctx, &newArticle)
			dbwrite += time.Now().UnixMilli() - start
			if err != nil {
				log.Println("Article with this title already exists!")
			}
		}
		data[section] = jsonRes
	}

	data["status"] = "OK"
	data["total"] = total
	data["write_time"] = dbwrite
	return data, nil
}
