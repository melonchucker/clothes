package models

type Browse struct {
	Items []struct {
		ItemName     string `json:"item_name"`
		BrandName    string `json:"brand_name"`
		Description  string `json:"description"`
		ThumbnailUrl string `json:"thumbnail_url"`
	} `json:"items"`
	TotalPages int `json:"total_pages"`
}
