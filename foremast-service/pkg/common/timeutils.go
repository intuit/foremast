package common

import (
	"log"
	"time"
)

// StrToTime .... convert input string format of time to time
func StrToTime(input string) time.Time {

	t1, err := time.Parse(
		time.RFC3339,
		input)
	if err != nil {
		log.Fatal(err)
		return time.Now().UTC()
	}
	return t1
}

//func main() {
//	var ts string = "2018-10-31T18:30:16-07:00"
//	fmt.Print(StrToTime(ts))
//}
