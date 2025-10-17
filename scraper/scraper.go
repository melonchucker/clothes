package scraper

import (
	"clothes/models"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gocolly/colly/v2"
)

func ScrapeAll() {
	slog.Info("Starting scraper")
	scrapeFashionPass()
}

type FashionPassResponse struct {
	Success     bool `json:"success"`
	ProductList struct {
		Pages            int               `json:"pages"`
		CurrentPage      int               `json:"current_page"`
		PageItemsTotal   int               `json:"page_items_total"`
		SearchItemsTotal int               `json:"search_items_total"`
		VendorList       map[string]string `json:"vendor_list"`
		ResultItems      []struct {
			ID                  string  `json:"id"`
			Title               string  `json:"title"`
			Handle              string  `json:"handle"`
			AverageReviewRating float64 `json:"averageReviewRating"`
			// They misspelled "product_cost" in their API
			PrductCost        float64         `json:"prduct_cost"`
			Retail            float64         `json:"retail"`
			Discount          float64         `json:"discount"`
			SaleStockDiscount float64         `json:"sale_stock_discount"`
			SalePrice         float64         `json:"sale_price"`
			NewItemDiscount   float64         `json:"newitem_discount"`
			UseItemDiscount   float64         `json:"useitem_discount"`
			Vendor            string          `json:"vendor"`
			VendorHandle      string          `json:"vendor_handle"`
			ThumbnailImage    string          `json:"thumbnail_image"`
			Images            []string        `json:"images"`
			Sizes             map[string]uint `json:"sizes"`
		} `json:"result_items"`
	} `json:"product_list"`
}

func scrapeFashionPass() {
	apiC := colly.NewCollector(
		colly.CacheDir("./.cache/fashionpass"),
	)

	apiC.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 1, Delay: 1 * time.Second})

	imgC := apiC.Clone()

	imgC.OnResponse(func(r *colly.Response) {
		slog.Info("Downloaded image", "url", r.Request.URL.String(), "size", len(r.Body))
		name := filepath.Base(r.Request.URL.Path)

		err := r.Save("./images/" + name)
		if err != nil {
			slog.Error("Failed to save image", "error", err)
		}
	})

	apiC.OnResponse(func(r *colly.Response) {
		j := FashionPassResponse{}
		json.Unmarshal(r.Body, &j)

		vendors := []string{}
		for _, v := range j.ProductList.VendorList {
			vendors = append(vendors, v)
		}
		_, err := models.GetDb().Exec(context.Background(), "INSERT INTO brand (name) SELECT unnest($1::text[]) ON CONFLICT (name) DO NOTHING", vendors)
		if err != nil {
			slog.Error("Failed to insert vendors", "error", err)
		}
		for _, item := range j.ProductList.ResultItems {
			images := []string{}
			for _, img := range item.Images {
				imgUrl := fmt.Sprintf("https://images.fashionpass.com/products/%s?profile=a", img)
				fmt.Println("Image URL:", imgUrl)
				images = append(images, imgUrl)
				imgC.Visit(imgUrl)
			}

			var baseID int64
			err := models.GetDb().QueryRow(context.Background(), `
				SELECT add_base_item($1, $2, $3, $4, $5, $6);
			`, item.Title, "", item.Vendor, item.ThumbnailImage, item.Images, item.AverageReviewRating).Scan(&baseID)
			if err != nil {
				slog.Error("Failed to insert item", "error", err, "item", item)
			}

			for size, count := range item.Sizes {
				var clothingID int64
				err := models.GetDb().QueryRow(context.Background(), `
				   SELECT add_clothing_item($1, $2);
				`, baseID, size).Scan(&clothingID)
				if err != nil {
					slog.Error("Failed to insert clothing item", "error", err, "item", item, "size", size)
				}

				_, err = models.GetDb().Exec(context.Background(), `
					SELECT api.transaction('audit', $1, $2);
				`, clothingID, count)
				if err != nil {
					slog.Error("Failed to insert stock audit", "error", err, "item", item, "size", size)
				}
			}
		}
	})

	apiC.OnRequest(func(r *colly.Request) {
		slog.Info("Visiting " + r.URL.String())
	})

	u := url.URL{}
	u.Host = "collections.fashionpass.com"
	u.Scheme = "https"
	u.Path = "/api/v1/collections/SearchByHandle2/clothing"
	q := u.Query()
	q.Set("items_per_page", "48")
	q.Set("sort_by", "pos")
	q.Set("sort_order", "desc")
	q.Set("page", "1")
	u.RawQuery = q.Encode()

	err := apiC.Visit(u.String())
	if err != nil {
		slog.Error("Error visiting site", "error", err)
		return
	}

	// http.Get("https://collections.fashionpass.com/api/v1/collections/SearchByHandle2/clothing?items_per_page=48&sort_by=pos&sort_order=desc&page=33&show_hidden_items=3&exclude_tags=bump-photo&flex_size=&default_size=&sort_by_size=false&in_stock=0&in_stock_sizes=0&isprice_for_customer=false&isSub=false&new_inStockFlag=true&auto_hide=true&is_customer_subscribed=false")
}
