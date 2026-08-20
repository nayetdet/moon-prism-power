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

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type callbackResult struct {
	Code  string
	State string
	Error string
}
