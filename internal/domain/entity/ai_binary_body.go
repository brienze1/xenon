package entity

import "mime"

func IsAIBinaryBody(binary bool, contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	return binary || mediaType == MediaTypeOctetStream
}
