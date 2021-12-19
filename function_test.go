package tools

import (
	"fmt"
	"testing"
)

func TestParseInt(t *testing.T) {
	num := ParseInt("1222.4")
	fmt.Println(num)
	num = ParseInt("这1243")
	fmt.Println(num)
	num = ParseInt("1243as")
	fmt.Println(num)
}

func TestTomorrow(t *testing.T) {
	fmt.Println(Tomorrow())
}
