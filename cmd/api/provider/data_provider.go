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

type Provider interface {
	GetArticles(context.Context) (map[string]interface{}, error)
}

type GuardianProvider struct {
	endpoint    string
	sections    []string
	queryParams map[string]string
	api_key     string
	catService  *service.CategoryService
	artService  *service.ArticleService
}

func NewGuardianProvider(endpoint string, sections []string, queryParams map[string]string, api_key string, cat *service.CategoryService, art *service.ArticleService) *GuardianProvider {
	return &GuardianProvider{
		endpoint:    endpoint,
		sections:    sections,
		queryParams: queryParams,
		api_key:     api_key,
		catService:  cat,
		artService:  art,
	}
}

//nytimes: politics, technology, science, sports, home, movies, business
//guardian: politics, technology, science, sport, -----, film, business

type Article struct {
	ApiURL             string `json:"apiUrl"`
	ID                 string `json:"id"`
	SectionName        string `json:"sectionName"`
	WebTitle           string `json:"webTitle"`
	WebPublicationDate string `json:"webPublicationDate"`
	WebUrl             string `json:"WebUrl"`
	Fields             Fields `json:"fields"`
	Tags               []Tag  `json:"tags"`
}

type Fields struct {
	Body      string `json:"body"`
	Thumbnail string `json:"thumbnail"`
}

type Tag struct {
	WebTitle string `json:"webTitle"`
}

type GuardianResponse struct {
	Response struct {
		Results []Article `json:"results"`
	} `json:"response"`
}

func (gp *GuardianProvider) GetArticles(ctx context.Context) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	formattedQueries := ""
	var dbwrite int64 = 0
	var total uint = 0

	fromDate, err := gp.artService.GetLatestTime(ctx, "The Guardian")
	toDate := time.Now().Format(time.DateOnly)

	if err == nil {
		formattedQueries += "&from-date=" + fromDate.Format(time.DateOnly)
		formattedQueries += "&to-date=" + toDate
	}

	for key, value := range gp.queryParams {
		formattedQueries += "&" + key + "=" + value
	}

	for _, section := range gp.sections {

		var response GuardianResponse

		res, err := http.Get(gp.endpoint + "?section=" + section + formattedQueries + "&api-key=" + gp.api_key + "&size=50")
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(body, &response); err != nil {
			return nil, err
		}

		for _, article := range response.Response.Results {

			total++

			category, err := gp.catService.GetByName(ctx, section)
			if err != nil {
				return nil, err
			}

			pubDate, _ := time.Parse(time.UnixDate, article.WebPublicationDate)

			newArticle := models.Article{
				Title:      article.WebTitle,
				Content:    article.Fields.Body,
				Source:     "The Guardian",
				ImageURL:   article.Fields.Thumbnail,
				CreatedAt:  pubDate,
				CategoryID: category.ID,
				Origin:     article.WebUrl,
			}
			start := time.Now().UnixMilli()
			err = gp.artService.Create(ctx, &newArticle)
			dbwrite += time.Now().UnixMilli() - start
			if err != nil {
				log.Println("Article with this title already exists!")
			}
		}
		data["status"] = "OK"
		data["total"] = total
		data["write_time"] = dbwrite
	}

	return data, nil
}
