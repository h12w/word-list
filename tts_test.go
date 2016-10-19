package main

import "testing"

func TestGetTTSToken(t *testing.T) {
	tk, err := getTTSToken("hello")
	if err != nil {
		t.Fatal(err)
	}
	if tk != "949705.540481" {
		t.Fatal("GetTTSToken Failed, got", tk)
	}
}
