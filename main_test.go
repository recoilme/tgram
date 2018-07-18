package main

import (
	"fmt"
	"testing"
)

func TestExtruct(t *testing.T) {
	for i := 1; i < 21; i++ {
		if i%10 == 0 {
			fmt.Println(i)
		}
	}
}
