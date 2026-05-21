package main

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type UpdateResponse struct {
	Result []Update `json:"result"`
}

type Listing struct {
	ID    int64
	Title string
	Price string
	URL   string
}

type NextData struct {
	Props struct {
		PageProps struct {
			Listings []struct {
				Listing struct {
					ID              int64  `json:"id"`
					Title           string `json:"title"`
					Price           string `json:"price"`
					SeoFriendlyPath string `json:"seoFriendlyPath"`
				} `json:"listing"`
			} `json:"listings"`
		} `json:"pageProps"`
	} `json:"props"`
}
