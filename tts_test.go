package main

import (
	"fmt"
	"testing"
)

func TestGetTTSToken(t *testing.T) {
	tk, err := getTTSToken("hello")
	if err != nil {
		t.Fatal(err)
	}
	if tk != "949705.540481" {
		t.Fatal("GetTTSToken Failed, got", tk)
	}
}

func TestGetTKK(t *testing.T) {
	fmt.Println(getTKK())
}
