package scanner

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

func shortMD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}

func ArtistID(name string) string {
	return "ar-" + shortMD5(strings.ToLower(name))
}

func AlbumID(artist, album string, year int64) string {
	return "al-" + shortMD5(fmt.Sprintf("%s|%s|%d", strings.ToLower(artist), strings.ToLower(album), year))
}

func SongID(path string) string {
	return "so-" + shortMD5(path)
}
