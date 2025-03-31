package services

import (
	"encoding/base64"
	"hash/fnv"
)

func GenerateShortKey(longURL string) string {
	h := fnv.New64a()
	h.Write([]byte(longURL))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
