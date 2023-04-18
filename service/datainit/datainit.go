package datainit

import (
	"errors"

	database "titable.go/db"
	// "titable.go/ical"
)

func BuildClassDatas(rowdata []database.Class) error {
	// icalenderからデータをパースした後のスライスから(X,Y,Length)を計算して更新する
	// 中途半端な開始時刻、終了時刻は規定の時限によるコマに丸める
	// 実装における仮定: 昼講義も56限の講義と同じ開始時刻, 日をまたぐ講義はない, 下2つの定数, 1限開始より前に終わらず、最終限終了より後に始まらない

	// 講義に関する定数2つ
	startTimes := [][]int{{8, 50}, {9, 40}, {10, 40}, {11, 30}, {14, 20}, {15, 10}, {16, 15}, {17, 5}, {18, 5}, {18, 55}} // 各時限の開始時刻
	const CLASSMINUTES = 50                                                                                               //1時限分の分数

	// 丸めのための補助変数, 分単位に統一
	var startStamps, endStamps []int
	for _, tmp := range startTimes {
		m := tmp[0]*60 + tmp[1]
		startStamps = append(startStamps, m)
		endStamps = append(endStamps, m+CLASSMINUTES)

	}

	for i, data := range rowdata {

		minuteStamp := data.Start.Hour()*60 + data.Start.Minute()
		switch {
		case data.Start.After(data.End):
			// 終了時刻のほうが開始時刻より前, assert
			return errors.New("error: invalid class time")
		case minuteStamp < startStamps[0]:
			// 1限より前
			data.Y = 0
		case startStamps[len(startStamps)-1] < minuteStamp:
			//startTimesの最後のコマより後に開始
			// エラー投げて終了せずにエラーになる講義コマを別途表示する感じにする？
			return errors.New("error: invalid class start time")
			// break
		case minuteStamp == startStamps[len(startStamps)-1]:
			data.Y = uint64(len(startStamps) - 1)
		default:
			// どこかでbreakできることが保証されてる
			for i := 0; i < len(startStamps)-1; i++ {
				if startStamps[i] <= minuteStamp && minuteStamp < startStamps[i+1] {
					data.Y = uint64(i)
					break
				}
			}
		}

		minuteStamp = data.End.Hour()*60 + data.End.Minute()
		switch {
		case minuteStamp <= endStamps[0]:
			data.Length = 1 - data.Y
		case endStamps[len(endStamps)-1] <= minuteStamp:
			data.Length = uint64(len(endStamps)) - data.Y
		default:
			for i := 0; i < len(endStamps)-1; i++ {
				if endStamps[i] < minuteStamp && minuteStamp <= endStamps[i+1] {
					data.Length = uint64(i+2) - data.Y
					break
				}
			}
		}

		data.X = uint64(data.Start.Weekday())

		rowdata[i] = data
	}
	return nil

}

/*
func main() {

	icalURL := "url"
	tmp,err := ical.GetCalData(icalURL)
	fmt.Println(err)
	BuildClassDatas(tmp)
	fmt.Println(tmp[0:5])

}
*/
