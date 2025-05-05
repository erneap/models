package pullinfo

type PullBook struct {
	Display     string `json:"display"`
	Code        string `json:"osis"`
	Testament   string `json:"testament"`
	NumChapters int    `json:"num_chapters"`
}
