package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
)

// word audio link:
// https://ssl.gstatic.com/dictionary/static/sounds/de/0/survey.mp3

func youdaoAudio(text string) (string, error) {
	return fmt.Sprintf(`http://dict.youdao.com/dictvoice?audio=%s&type=2`, text), nil
}

func googleAudioURL(text string) (string, error) {
	return fmt.Sprintf(
		`https://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&client=tw-ob&q=%s&tl=en_US`,
		text,
	), nil
}

func googleSignedAudioURL(text string) (string, error) {
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

var rxTKK = regexp.MustCompile(`TKK=eval\('\(\(function\(\){var a\\x3d([0-9]+);var b\\x3d([0-9]+);return ([0-9]+)`)

func getTKK() (x, a, b string, err error) {
	req, err := http.NewRequest("GET", "https://translate.google.com/", nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}
	m := rxTKK.FindSubmatch(buf)
	if len(m) != 4 {
		return "", "", "", errors.New("fail to find TKK")
	}
	return string(m[3]), string(m[1]), string(m[2]), nil
}

func getTTSToken(text string) (string, error) {
	x, a, b, err := getTKK()
	if err != nil {
		return "", err
	}
	out, err := exec.Command("js/gen_tk.js", text, x, a, b).Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(out)), nil
}
