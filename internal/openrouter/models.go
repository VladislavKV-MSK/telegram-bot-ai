package openrouter

// Message представляет одно сообщение в диалоге
type Message struct {
	Role    string `json:"role"`    // "user" или "assistant"
	Content string `json:"content"` // Текст сообщения
}

// Request представляет запрос к OpenRouter API
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Provider struct {
		AllowFallbacks bool     `json:"allow_fallbacks"`
		Order          []string `json:"order"`
		MaxPrice       struct {
			Prompt     float64 `json:"prompt"`
			Completion float64 `json:"completion"`
		} `json:"max_price"`
		RequireParameters bool `json:"require_parameters"`
	} `json:"provider"`
}

// Response представляет ответ от OpenRouter API
type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
