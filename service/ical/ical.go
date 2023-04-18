package ical

import (
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	database "titable.go/db"
)

func GetCalData(icalURL string) ([]database.Class, error) {
	res, err := http.Get(icalURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	cal, err := ics.ParseCalendar(res.Body)
	if err != nil {
		return nil, err
	}

	classes := []database.Class{}
	jst, _ := time.LoadLocation("Asia/Tokyo")
	for _, event := range cal.Components {
		var c database.Class
		for _, eventProperty := range event.UnknownPropertiesIANAProperties() {
			switch eventProperty.IANAToken {
			case "SUMMARY":
				c.Class = eventProperty.Value
			case "DTSTART":
				c.Start, _ = time.ParseInLocation("20060102T150405", eventProperty.Value, jst)
			case "DTEND":
				c.End, _ = time.ParseInLocation("20060102T150405", eventProperty.Value, jst)
				/* Class.Commentに場所、説明（サブタイトルなど）を追加します。不要ならコメントアウト
				case "LOCATION":
					if eventProperty.Value != "" {
						c.Comment = "Location:" + eventProperty.Value + ", " + c.Comment
					}
				case "DESCRIPTION":
					if !regexp.MustCompile("　|この講義日程は自動*").MatchString(eventProperty.Value) {
						c.Comment = c.Comment + "Description:" + eventProperty.Value
					}
				//*/
			}
		}
		if c.Class != "" { // これがないと最初に変なのが入る
			classes = append(classes, c)
		}
	}

	return classes, nil
}
