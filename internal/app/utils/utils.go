package utils

import (
	"encoding/base64"
	"hash/fnv"
)

func GenerateShortKey(longUrl string) string {
	h := fnv.New64a()
	h.Write([]byte(longUrl))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
