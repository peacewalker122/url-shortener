package http

type CreateShortURLRequest struct {
	URL string `json:"url"`
}

type ShortURLResponse struct {
	URL string `json:"url"`
}

type ErrorResponse struct {
	Detail string `json:"detail"`
}
