package utils

import (
	"regexp"
	"strings"
)

var vietnameseMap = map[rune]string{
	'à': "a", 'á': "a", 'ả': "a", 'ã': "a", 'ạ': "a",
	'ă': "a", 'ằ': "a", 'ắ': "a", 'ẳ': "a", 'ẵ': "a", 'ặ': "a",
	'â': "a", 'ầ': "a", 'ấ': "a", 'ẩ': "a", 'ẫ': "a", 'ậ': "a",
	'đ': "d",
	'è': "e", 'é': "e", 'ẻ': "e", 'ẽ': "e", 'ẹ': "e",
	'ê': "e", 'ề': "e", 'ế': "e", 'ể': "e", 'ễ': "e", 'ệ': "e",
	'ì': "i", 'í': "i", 'ỉ': "i", 'ĩ': "i", 'ị': "i",
	'ò': "o", 'ó': "o", 'ỏ': "o", 'õ': "o", 'ọ': "o",
	'ô': "o", 'ồ': "o", 'ố': "o", 'ổ': "o", 'ỗ': "o", 'ộ': "o",
	'ơ': "o", 'ờ': "o", 'ớ': "o", 'ở': "o", 'ỡ': "o", 'ợ': "o",
	'ù': "u", 'ú': "u", 'ủ': "u", 'ũ': "u", 'ụ': "u",
	'ư': "u", 'ừ': "u", 'ứ': "u", 'ử': "u", 'ữ': "u", 'ự': "u",
	'ỳ': "y", 'ý': "y", 'ỷ': "y", 'ỹ': "y", 'ỵ': "y",
	'À': "a", 'Á': "a", 'Ả': "a", 'Ã': "a", 'Ạ': "a",
	'Ă': "a", 'Ằ': "a", 'Ắ': "a", 'Ẳ': "a", 'Ẵ': "a", 'Ặ': "a",
	'Â': "a", 'Ầ': "a", 'Ấ': "a", 'Ẩ': "a", 'Ẫ': "a", 'Ậ': "a",
	'Đ': "d",
	'È': "e", 'É': "e", 'Ẻ': "e", 'Ẽ': "e", 'Ẹ': "e",
	'Ê': "e", 'Ề': "e", 'Ế': "e", 'Ể': "e", 'Ễ': "e", 'Ệ': "e",
	'Ì': "i", 'Í': "i", 'Ỉ': "i", 'Ĩ': "i", 'Ị': "i",
	'Ò': "o", 'Ó': "o", 'Ỏ': "o", 'Õ': "o", 'Ọ': "o",
	'Ô': "o", 'Ồ': "o", 'Ố': "o", 'Ổ': "o", 'Ỗ': "o", 'Ộ': "o",
	'Ơ': "o", 'Ờ': "o", 'Ớ': "o", 'Ở': "o", 'Ỡ': "o", 'Ợ': "o",
	'Ù': "u", 'Ú': "u", 'Ủ': "u", 'Ũ': "u", 'Ụ': "u",
	'Ư': "u", 'Ừ': "u", 'Ứ': "u", 'Ử': "u", 'Ữ': "u", 'Ự': "u",
	'Ỳ': "y", 'Ý': "y", 'Ỷ': "y", 'Ỹ': "y", 'Ỵ': "y",
}

var (
	reNonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)
	reTrimDash    = regexp.MustCompile(`^-+|-+$`)
)

// Slugify converts a Vietnamese (or Latin) string into a URL-safe slug.
// "Du lịch biển" → "du-lich-bien"
func Slugify(s string) string {
	var b strings.Builder
	for _, r := range s {
		if rep, ok := vietnameseMap[r]; ok {
			b.WriteString(rep)
		} else {
			b.WriteRune(r)
		}
	}
	slug := strings.ToLower(b.String())
	slug = reNonAlphaNum.ReplaceAllString(slug, "-")
	slug = reTrimDash.ReplaceAllString(slug, "")
	return slug
}
