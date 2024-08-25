package dify

import "regexp"

func ExtractURLFromImageString(input string) *string {
	// 正規表現パターン
	pattern := `!\[image\]\(([^)]+)\)`
	re := regexp.MustCompile(pattern)

	// 正規表現にマッチするかチェック
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return &matches[1]
	}

	return nil
}
