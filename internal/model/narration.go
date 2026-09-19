package model

type Narration struct {
	Episode    int                 `json:"episode"`
	Title      map[string]*string  `json:"title"`
	Narrations map[string][]string `json:"narrations"`
}
