package main

import (
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

func readBook(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return string(content), err
	}

	return string(content), nil
}

func TestParse(t *testing.T) {
	book, err := readBook("t/000.txt")
	if err != nil {
		log.Fatalln("[error]", err)
	}

	debug := false
	header, body, _ := parse(&debug, book)

	headert := reflect.TypeOf(header).Kind()
	bodyt := reflect.TypeOf(body).Kind()

	if headert != reflect.Slice {
		t.Errorf("got %s, expected %s", headert, reflect.Slice)
	}

	if bodyt != reflect.Slice {
		t.Errorf("got %s, expected %s", bodyt, reflect.Slice)
	}
}
