package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"h12.me/html-query"
	"h12.me/html-query/expr"
)

func main() {
	cache := newCache()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`
<html>
<head>
<title></title>
</head>
<body>
<form action="/" method="post">
    <textarea name="words" cols="80" rows="10"></textarea>
    <input type="submit" value="Go">
</form>
</body>
</html>
`))
		case "POST":
			r.ParseForm()
			words := splitWords(r.Form.Get("words"))
			wordParams := strings.Join(words, ".")
			uri := "/word/?w=" + wordParams
			http.Redirect(w, r, uri, http.StatusFound)
		}
	})
	http.HandleFunc("/word/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		debug := q.Get("debug") == "1"
		words := strings.Split(q.Get("w"), ".")
		index, err := strconv.Atoi(q.Get("i"))
		if err != nil {
			index = 0
		}
		if index >= 0 && index < len(words) {
			word := words[index]
			images, err := googleImages(cache, word)
			if err != nil {
				if debug {
					w.Write([]byte(err.Error()))
					return
				}
			}
			definition, audio, err := googleDefinition(cache, word)
			if err != nil {
				if debug {
					w.Write([]byte(err.Error()))
					return
				}
			}

			// HTTP Start
			w.Write([]byte("<html><head></head><body>"))

			// Navigation
			w.Write([]byte("<div>"))
			if index > 0 {
				q.Set("i", strconv.Itoa(index-1))
				w.Write([]byte(fmt.Sprintf(`<a href="%s">Prev</a>`, "/word/?"+q.Encode())))
			}
			if index < len(words)-1 {
				q.Set("i", strconv.Itoa(index+1))
				w.Write([]byte(fmt.Sprintf(`<a href="%s">Next</a>`, "/word/?"+q.Encode())))
			}
			w.Write([]byte("</div>"))

			// Title
			w.Write([]byte("<h1>" + word + "</h1>"))

			// Audio
			w.Write([]byte(fmt.Sprintf(`
				<div><audio controls>
				  <source src="%s" type="audio/mpeg">
				</audio></div>`, audio)))

			// Pictures
			for _, img := range images {
				w.Write([]byte("<img src=" + img + "></img>"))
			}

			// Definition
			w.Write([]byte(definition))

			// HTTP End
			w.Write([]byte("</body></html>"))
		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func googleImages(client Getter, word string) ([]string, error) {
	var (
		Id  = expr.Id
		Img = expr.Img
	)
	req, err := http.NewRequest("GET", "https://www.google.com/search?safe=strict&tbm=isch&q="+word, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", `Mozilla/5.0 (Android 4.4; Mobile; rv:41.0) Gecko/41.0 Firefox/41.0`)
	body, err := client.Get(req)
	if err != nil {
		return nil, err
	}
	root, err := query.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	images := root.Div(Id("search")).Descendants(Img).Strings(expr.GetSrc)
	if len(images) == 0 {
		err := errors.New("cannot find images")
		if page := root.Render(); page != nil {
			err = errors.New(*page)
		}
		return nil, err
	}
	return images, nil
}

func googleDefinition(client Getter, word string) (def, audio string, err error) {
	var (
		Class = expr.Class
	)
	req, err := http.NewRequest("GET", "https://www.google.com/search?gws_rd=cr,ssl&hl=en&q=define%3A"+word, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.59 Safari/537.36`)
	body, err := client.Get(req)
	if err != nil {
		return "", "", err
	}
	root, err := query.Parse(bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	dict := root.Div(Class("lr_dct_ent"))
	if dictHTML := dict.Render(); dictHTML != nil {
		if audioLink := dict.Audio().Src(); audioLink != nil {
			audio = *audioLink
		}
		def = *dictHTML
	}
	if def == "" {
		err := errors.New("cannot find definition")
		if page := root.Render(); page != nil {
			err = errors.New(*page)
		}
		return "", "", err

	}
	return
}

var rxSpace = regexp.MustCompile(`[\t \r\n!()\[\]\{\};:",<.>?“”‘’*/]+`)

func splitWords(s string) []string {
	words := rxSpace.Split(s, -1)
	m := make(map[string]bool)
	var results []string
	for _, word := range words {
		if _, exists := m[word]; !exists {
			m[word] = true
			results = append(results, word)
		}
	}
	return words
}

var (
	client = http.Client{}
)

type cache struct {
	m map[string][]byte
}

func newCache() *cache {
	return &cache{
		m: make(map[string][]byte),
	}
}

type Getter interface {
	Get(req *http.Request) ([]byte, error)
}

func (c *cache) Get(req *http.Request) ([]byte, error) {
	uri := req.URL.String()
	if body, ok := c.m[uri]; ok {
		return body, nil
	}
	req.Header.Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8`)
	req.Header.Set("Accept-Language", "en-US,en")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.m[req.URL.String()] = body
	return body, nil
}