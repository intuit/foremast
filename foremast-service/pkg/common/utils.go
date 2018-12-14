package common

import (
	"bytes"
	"fmt"
	"strings"
)

func ConvertMapToString(m map[string]string) string {
	b := new(bytes.Buffer)
	size := len(m)
	i := 1
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%s\"", key, value)
		if i < size {
			fmt.Fprintf(b, ",")
		}
		i++

	}
	return b.String()
}

func ConvertStringToMap(str string) map[string]string {
	m := make(map[string]string)
	strs := strings.Split(str, ",")
	for _, s := range strs {
		if len(s) != 0 {
			sm := strings.Split(s, "=")
			m[sm[0]] = sm[1]
		}
	}
	return m
}

/*
func main1() {
	m := make(map[string]string)
	m["k1"] = "ab"
	m["k2"] = "cd"
	ss := ConvertMapToString(m)

	mm := ConvertStringToMap(ss)
	fmt.Println(" mm ", mm)
}*/
