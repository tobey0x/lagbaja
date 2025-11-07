package models

type Flashcard struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Topic    string `json:"topic,omitempty"`
}

type FlashcardSet struct {
	Title      string      `json:"title"`
	Flashcards []Flashcard `json:"flashcards"`
	Source     string      `json:"source"`
	CreatedAt  string      `json:"createdAt"`
	TotalCards int         `json:"totalCards"`
}

type PDFProcessRequest struct {
	URL      string
	FilePath string
	Content  []byte
}