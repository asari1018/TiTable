package service

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	database "titable.go/db"
)

// (カレンダーがある)メイン画面のサーバ側

func Main(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからユーザIDを得る
	cookie, err := ctx.Request.Cookie("id")
	uid, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	// Get tasks in DB
	var tasks []database.Task
	query := "SELECT * FROM tasks WHERE is_done=false AND id IN (SELECT task_id FROM user_info WHERE user_id=" + uid + ")"
	err = db.Select(&tasks, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println(tasks)

	// Get classes in DB
	var date time.Time // 表示するカレンダーの期間(1週間)の一日
	request_date := ctx.Param("date")
	jst, _ := time.LoadLocation("Asia/Tokyo")
	if request_date != "" {
		date, _ = time.ParseInLocation("20060102", request_date, jst)
	} else {
		date = time.Now()
	}
	classes, err := getWeeklySchedule(uid, date)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	// カレンダーの期間
	yearMonth := date.Format("2006年1月")
	var dateRange []string
	for i := 0; i < 7; i++ {
		dateRange = append(dateRange, date.AddDate(0, 0, i-int(date.Weekday())).Format("2日"))
	}

	// Render tasks
	ctx.HTML(http.StatusOK, "main.html",
		gin.H{
			"Title":     "TITABLE",
			"Tasks":     tasks,
			"Classes":   classes,
			"YearMonth": yearMonth,
			"LastWeek":  date.AddDate(0, 0, -7).Format("20060102"),
			"NextWeek":  date.AddDate(0, 0, 7).Format("20060102"),
			"DateRange": dateRange,
			"CalBody":   template.HTML(makeCalendarBody(classes)),
		},
	)
}

// 引数time time.Timeを含む1週間の予定を[]database.Classで取得
func getWeeklySchedule(uid string, time time.Time) ([]database.Class, error) {
	db, err := database.GetConnection()
	if err != nil {
		return nil, err
	}

	var classes []database.Class
	const format = "2006-01-02"
	date0 := time.AddDate(0, 0, 0-int(time.Weekday())) // Sunday
	date6 := time.AddDate(0, 0, 6-int(time.Weekday())) // Saturday
	query := fmt.Sprintf("SELECT * FROM classes WHERE uid=%s AND start BETWEEN '%s 00:00:00' AND '%s 23:59:59'", uid, date0.Format(format), date6.Format(format))
	err = db.Select(&classes, query)
	if err != nil {
		return nil, err
	}

	return classes, nil
}

// カレンダーのボディを作ります。htmlテンプレートで<table>タグとその中の<tbody>タグで括られる前提
func makeCalendarBody(classes []database.Class) string {
	var cal string
	const t = "  " // HTMLのindent
	// Class.Y, Class.Xの優先順位で辞書順ソート
	sort.SliceStable(classes, func(i, j int) bool { return classes[i].X < classes[j].X })
	sort.SliceStable(classes, func(i, j int) bool { return classes[i].Y < classes[j].Y })
	// カレンダーのボディを作る
	i := 0                             // 今考える授業のインデックス
	pass := []int{0, 0, 0, 0, 0, 0, 0} // rowspanで消費した分の下のカラムは書かない
	for row := 0; row < 10; row++ {    // 1限から10限まで
		cal += "<tr class=body>\n"
		cal += fmt.Sprintf("%s<th>%d</th>\n", t, row+1)
		for col := 0; col < 7; col++ { // 日曜から土曜まで
			if pass[col] > 0 {
				pass[col]--
			} else if i < len(classes) && int(classes[i].Y) == row && int(classes[i].X) == col {
				cal += fmt.Sprintf("%s<td class=\"exit_class\" rowspan=%d>\n", t, classes[i].Length)
				//cal += fmt.Sprintf("%s%s%s\n", t, t, classes[i].Class)
				cal += fmt.Sprintf("%s%s<a href=\"/class/%d\">%s</a>\n", t, t, classes[i].ID, classes[i].Class)
				cal += fmt.Sprintf("%s</td>\n", t)
				pass[col] = int(classes[i].Length) - 1
				i++
			} else {
				cal += fmt.Sprintf("%s<td>&nbsp;</td>\n", t)
			}
		}
		cal += "</tr>\n"
	}
	return cal
}
