package provider

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Blagoja0123/skippy/cmd/api/service"
	"github.com/Blagoja0123/skippy/models"
)

type NewsProvider struct {
	endpoint    string
	sections    []string
	queryParams map[string]string
	api_key     string
	catService  service.CategoryService
	artService  service.ArticleService
}

func NewNewsProvider(
	endpoint string,
	sections []string,
	queryParams map[string]string,
	api_key string,
	cat service.CategoryService,
	art service.ArticleService,
) *NewsProvider {
	return &NewsProvider{
		endpoint:    endpoint,
		sections:    sections,
		queryParams: queryParams,
		api_key:     api_key,
		catService:  cat,
		artService:  art,
	}
}

type newsResponse struct {
	Status  string        `json:"status"`
	Results []newsArticle `json:"news"`
}

type newsArticle struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"image"`
	WebURL   string `json:"url"`
	PubDate  string `json:"published`
}

func (np *NewsProvider) GetArticles(ctx context.Context) (map[string]interface{}, error) {

	data := make(map[string]interface{})

	formattedParams := ""

	for key, value := range np.queryParams {
		formattedParams += "&" + key + "=" + value
	}
	var dbwrite int
	var start, end time.Time
	total := 0

	fromData, err := np.artService.GetLatestTime(ctx, "Currents")
	if err == nil {
		formattedParams += "&start_date=" + fromData.Format(time.RFC3339)
		formattedParams += "&end_date=" + time.Now().Format(time.RFC3339)
	}

	for _, section := range np.sections {
		log.Println("Looking through sections!")
		res, err := http.Get(np.endpoint + "?apiKey=" + np.api_key + "&category=" + section + formattedParams)
		if err != nil {
			return nil, err
		}

		body, _ := io.ReadAll(res.Body)
		var jsonRes newsResponse
		var rawRes interface{}
		if err := json.Unmarshal(body, &jsonRes); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(body, &rawRes); err != nil {
			return nil, err
		}

		for _, result := range jsonRes.Results {
			total++
			category, err := np.catService.GetByName(ctx, section)

			if err != nil {
				return nil, err
			}

			pubDate, _ := time.Parse(time.UnixDate, result.PubDate)

			newArticle := models.Article{
				Title:      result.Title,
				Content:    result.Content,
				Source:     "Currents",
				Origin:     result.WebURL,
				CreatedAt:  pubDate,
				ImageURL:   result.ImageURL,
				CategoryID: category.ID,
			}

			start = time.Now()
			err = np.artService.Create(ctx, &newArticle)
			end = time.Now()
			dbwrite += int(end.UnixMilli()) - int(start.UnixMilli())
			if err != nil {
				return nil, err
			}
		}
		data[section] = rawRes
	}

	data["success"] = true
	data["write_time"] = strconv.Itoa(dbwrite) + " ms"
	data["total"] = total
	return data, nil
}
