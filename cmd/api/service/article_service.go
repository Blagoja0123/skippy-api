package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Blagoja0123/skippy/models"
	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(DB *gorm.DB) *ArticleService {
	return &ArticleService{
		db: DB,
	}
}

func (as *ArticleService) Get(ctx context.Context, params map[string]string) ([]models.Article, error) {

	var articles []models.Article

	query := as.db.WithContext(ctx).Model(&models.Article{}).Preload("Category")

	if value, exists := params["category_id"]; exists {
		splitCats := strings.Split(value, ",")
		query.Where("category_id IN ?", splitCats)
	}

	if value, exists := params["within_last"]; exists {
		last := time.Now()

		sufStripped := strings.Clone(value)
		sufStripped = value[:len(sufStripped)-1]
		duration, err := strconv.Atoi(sufStripped)

		if err != nil {
			return nil, err
		}
		switch {
		case strings.HasSuffix(value, "m"):
			last = last.AddDate(0, -duration, 0)
		case strings.HasSuffix(value, "w"):
			last = last.AddDate(0, 0, -duration*7)
		case strings.HasSuffix(value, "d"):
			last = last.AddDate(0, 0, -duration)
		case strings.HasSuffix(value, "h"):
			last = last.Add(-time.Duration(duration) * time.Hour)
		default:
			log.Println("Invalid suffix in within_last parameter")
			return nil, fmt.Errorf("invalid suffix in within_last parameter")
		}

		query.Where("created_at >= ?", last)
	}

	pageLimit := 15

	if value, exists := params["limit"]; exists {
		pageLimit, _ = strconv.Atoi(value)
	}

	if value, exists := params["source"]; exists {
		query.Where("source = ?", value)
	}

	if value, exists := params["page"]; exists {
		value, _ := strconv.Atoi(value)
		offset := value - 1*pageLimit
		query.Offset(offset)
	}

	if err := query.Order("RANDOM()").Limit(pageLimit).Find(&articles).Error; err != nil {
		return nil, err
	}

	return articles, nil
}

func (as *ArticleService) GetLatestTime(ctx context.Context, source string) (time.Time, error) {

	var article models.Article
	if err := as.db.WithContext(ctx).Model(&models.Article{}).Where("source = ?", source).Order("created_at DESC").First(&article).Error; err != nil {
		return time.Now(), err
	}

	return article.CreatedAt, nil
}

func (as *ArticleService) GetByID(ctx context.Context, id int) (*models.Article, error) {
	var article models.Article

	if err := as.db.WithContext(ctx).First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}

func (as *ArticleService) Create(ctx context.Context, body *models.Article) error {
	if body.Source == "The New York Times" && body.ImageURL == "" {
		body.ImageURL = "https://help.nytimes.com/hc/theming_assets/01HZPCK5BKMK9ZRNEE1Y6J1PHW"
	}
	if body.Source == "The Guardian" && body.ImageURL == "" {
		body.ImageURL = "https://p7.hiclipart.com/preview/707/137/261/the-guardian-guardian-media-group-theguardian-com-news-journalism-the-guardian-logo.jpg"
	}
	return as.db.WithContext(ctx).Model(&models.Article{}).Create(body).Error
}

func (as *ArticleService) Update(ctx context.Context, body *models.Article) error {
	return as.db.WithContext(ctx).Model(&models.Article{}).Save(body).Error
}

func (as *ArticleService) Delete(ctx context.Context, id int) error {
	return as.db.WithContext(ctx).Model(&models.Article{}).Delete(&models.Article{}, id).Error
}

func (as *ArticleService) BulkDelete(ctx context.Context, source string) error {
	return as.db.WithContext(ctx).Model(&models.Article{}).Where("source = ?", source).Delete(&models.Article{}).Error
}
