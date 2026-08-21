package mal

type listResponse struct {
	Data   []listEntry `json:"data"`
	Paging listPaging  `json:"paging"`
}

type listEntry struct {
	Node listNode `json:"node"`
}

type listNode struct {
	ID int `json:"id"`
}

type listPaging struct {
	Next string `json:"next"`
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type callbackResult struct {
	Code  string
	State string
	Error string
}
