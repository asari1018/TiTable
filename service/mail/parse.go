package mail

import (
	"errors"
	"regexp"
)

// 文字列からurlを取得
func FetchURL(text string) (string, error) {
	var url string
	err := errors.New("url does not exist")

	arr := regexp.MustCompile("\r\n|\n").Split(text, -1) // 改行コードの正規表現で文字列をスプリット
	regHttp := regexp.MustCompile("https://*|http://*")
	regZoom := regexp.MustCompile("zoom")
	regShare := regexp.MustCompile("rec/share")
	for i := 0; i < len(arr); i++ {
		if regHttp.MatchString(arr[i]) && regZoom.MatchString(arr[i]) && !regShare.MatchString(arr[i]) {
			url = arr[i]
			err = nil
		}
	}

	return url, err
}
