package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func audioURL(text string) (string, error) {
	tk, err := getTTSToken(text)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`https://translate.google.com/translate_tts?ie=UTF-8&q=%s&tl=en&total=1&idx=0&textlen=%d&tk=%s&client=t&prev=input`,
		text,
		len(text),
		tk,
	), nil
}

func getTTSToken(text string) (string, error) {
	out, err := exec.Command("js/gen_tk.js", text).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}
