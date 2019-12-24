package main_test

import (
	"fmt"
	"testing"
	"time"
)

func Test1(tt *testing.T) {
	t, _ := time.Parse("20060102", "20191119")

	fmt.Println(t)
}
