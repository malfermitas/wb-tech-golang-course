package responses

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type EventResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
