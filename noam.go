package main
import (
	//"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var file_name = "/Users/noamsohn/Desktop/shorten.csv"

	csvfile, err := os.Open(file_name)
	if err != nil {
		log.Fatalln("couldnt open the csv file", err)
	}

	r := csv.NewReader(csvfile)

	for {
		record, err :=  r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("the type is %T\n", record[0])
		//fmt.Printf("The record is: %s\n", record[0])

	}
}
