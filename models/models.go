package models

type Brands map[string][]string

func (b Brands) Letters() []string {
	return []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "#"}
}

type Browse struct {
	Items []struct {
		ItemName     string `json:"item_name"`
		BrandName    string `json:"brand_name"`
		Description  string `json:"description"`
		ThumbnailUrl string `json:"thumbnail_url"`
		BaseItemID   int    `json:"base_item_id"`
	} `json:"items"`
	TotalPages int `json:"total_pages"`
}

type Detail struct {
	ItemName            string   `json:"item_name"`
	BrandName           string   `json:"brand_name"`
	Rating              float64  `json:"rating"`
	Description         string   `json:"description"`
	ImageUrls           []string `json:"image_urls"`
	ThumbnailUrl        string   `json:"thumbnail_url"`
	ItemSpecificDetails []struct {
		Size    string `json:"size"`
		InStock bool   `json:"in_stock"`
	} `json:"item_specific_details"`
	Tags     []string `json:"tags"`
	MoreLike struct {
		SameBrand []struct {
			ItemName     string `json:"item_name"`
			BrandName    string `json:"brand_name"`
			ThumbnailUrl string `json:"thumbnail_url"`
		} `json:"more_from_brand"`
		SimilarItems []struct {
			ItemName     string `json:"item_name"`
			BrandName    string `json:"brand_name"`
			ThumbnailUrl string `json:"thumbnail_url"`
		} `json:"similar_items"`
	} `json:"more_like"`
}

type SearchBar struct {
	Tags   []string `json:"tags"`
	Brands []string `json:"brands"`
	Items  []string `json:"items"`
}

type SiteUser struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IsStaff   bool   `json:"is_staff"`
	IsAdmin   bool   `json:"is_admin"`
}
