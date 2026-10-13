package dto

const GitProgressEvent = "git:progress"

type GitProgress struct {
	LocalId string `json:"local_id"`
	Phase   string `json:"phase"`
	Done    int    `json:"done"`
	Total   int    `json:"total"`
}

func NewGitProgress(localId, phase string, done, total int) GitProgress {
	return GitProgress{LocalId: localId, Phase: phase, Done: done, Total: total}
}
