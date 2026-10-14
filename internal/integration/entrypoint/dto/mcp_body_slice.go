package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type MCPBodySlice struct {
	Text      string `json:"text"`
	Offset    int    `json:"offset"`
	Size      int    `json:"size"`
	Truncated bool   `json:"truncated,omitempty"`
}

func NewMCPBodySlice(slice entity.AIBodySlice) MCPBodySlice {
	return MCPBodySlice{Text: slice.Text, Offset: slice.Offset, Size: slice.Size, Truncated: slice.Truncated}
}
