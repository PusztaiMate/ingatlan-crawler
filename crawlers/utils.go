package crawlers

import (
	"encoding/csv"
	"log"
	"os"
	"strings"
)

func JoinUri(a, b string) string {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return b
	}

	a = strings.TrimRight(a, "/")
	b = strings.TrimLeft(b, "/")

	return a + "/" + b
}

func WritePropertiesToCsv(filepath string, props []PropertyInfo) {
	var propsWritable [][]string
	propsWritable = append(propsWritable, PropertyInfo{}.GetHeaders())
	for _, prop := range props {
		propsWritable = append(propsWritable, prop.ToSlice())
	}

	f, err := os.Create(filepath)
	if err != nil {
		log.Fatalf("could not create file '%s', can not save data", filepath)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	err = writer.WriteAll(propsWritable)
	if err != nil {
		log.Fatalln("could not write csv ")
	}

}
