package domain

type GitHubPushEvent struct {
	Ref        string `json:"ref"` // เพิ่ม field นี้
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		HTMLURL  string `json:"html_url"` // เพิ่ม field นี้
	} `json:"repository"`
	Commits []struct {
		Message string `json:"message"`
		Author  struct {
			Name string `json:"name"`
		} `json:"author"`
	} `json:"commits"`
}
