package main

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"h12.me/html-query"
	"h12.me/html-query/expr"
	"h12.me/socks"
)

var (
	client = http.Client{
		Transport: &http.Transport{
			Dial:                  socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080"),
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

const (
	ua = `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.59 Safari/537.36`
	// ua = `Lynx/2.8.5rel.1 libwww-FM/2.14 SSL-MM/1.4.1 GNUTLS/1.0.16`
)

var filteredKeys = map[string]bool{
	"X-Frame-Options":  true,
	"X-Xss-Protection": true,
}

func main() {
	// cache := newCache()
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
			uri := "/words/?w=" + wordParams
			http.Redirect(w, r, uri, http.StatusFound)
			// go prefetch(cache, words)
		}
	})
	http.HandleFunc("/word/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		word := q.Get("w")

		tmpl, err := template.New("words").Parse(`
		<!DOCTYPE html>
		<html lang="es-US">
		<head>
		<meta name="language" content="English">
		</head>
		<style>
		#image_container {
			width    : 1200px;
			height   : 350px;
			overflow : hidden;
			position : relative;
		}
		#image
		{
			position : relative;
			top      : -150px;
			left     : 0px;
			width    : 1200px;
			height   : 500px;
		}
		#dict_container {
			margin-top: 50px;
			width    : 800px;
			height   : 600px;
			overflow : hidden;
			position : relative;
		}
		#dict
		{
			position : relative;
			top      : -230px;
			left     : -150px;
			width    : 800px;
			height   : 600px;
		}
		</style>
		<body>
		<div id="image_container">
		<iframe id="image" src="https://www.google.com/search?safe=strict&tbm=isch&q={{.word}}"></iframe>
		</div>
		<div id="dict_container">
		<iframe id="dict" src="https://www.google.com/search?gws_rd=cr&hl=en&q=define%3A{{.word}}"></iframe>
		</div>
		</body>
		</html>
				`)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
		tmpl.Execute(w, map[string]interface{}{
			"word": word,
		})

	})
	http.HandleFunc("/dict/", func(w http.ResponseWriter, r *http.Request) {
		word := r.URL.Query().Get("w")
		req, err := http.NewRequest("GET", "https://www.google.com/search?gws_rd=cr&hl=en&q=define%3A"+word, nil)
		// req, err := http.NewRequest("GET", "https://translate.google.com/#en/zh-CN/"+word, nil)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := client.Transport.RoundTrip(req)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			if !filteredKeys[k] {
				w.Header()[k] = v
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
	http.HandleFunc("/words/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		// debug := q.Get("debug") == "1"
		words := strings.Split(q.Get("w"), ".")
		// index, err := strconv.Atoi(q.Get("i"))
		// if err != nil {
		// index = 0
		// }

		tmpl, err := template.New("words").Parse(`
		<!DOCTYPE html>
		<html lang="es-US">
		<head>
		<meta name="language" content="English">
		</head>
		
		<style>
		.links > a {
		  display:block;
		}
		#left {
			float: left;
			width: 250px;
			height: 100%;
		}
		#right {
						position: fixed;
						top: 0;
						bottom: 0;
						left: 250px;
						width: 100%;
						overflow-y: hidden;
		}
		</style>
		
		<body>
		<div id="left" class="links">
		{{range $word := .Words}}
		<a href="/word/?w={{$word}}" target="word-iframe">{{$word}}</a>
		{{end}}
		</div>
		
		<iframe id="right" name="word-iframe" height="100%"></iframe>
		</div>
		
		</body>
		</html>
				`)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Words": words,
		})

		/*
			if index >= 0 && index < len(words) {
				word := words[index]
				images, err := googleImages(cache, word)
				if err != nil {
					if debug {
						w.Write([]byte(err.Error()))
						return
					}
				}
				definition, err := googleDefinition(cache, word)
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
				audio := `/tts?q=` + word
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
		*/
	})

	/*
		http.HandleFunc("/tts", func(w http.ResponseWriter, r *http.Request) {
			word := r.URL.Query().Get("q")
			uri, _ := googleAudioURL(word)
			req, err := http.NewRequest("GET", uri, nil)
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("User-Agent", ua)
			req.Header.Set("Referer", "http://translate.google.com/")
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			defer resp.Body.Close()
			w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
			io.Copy(w, resp.Body)
		})
	*/

	log.Fatal(http.ListenAndServe(":7677", nil))
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
	req.Header.Set("User-Agent", ua)
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

func googleDefinition(client Getter, word string) (string, error) {
	var (
		Class = expr.Class
	)
	req, err := http.NewRequest("GET", "https://www.google.com/search?gws_rd=cr,ssl&hl=en&q=define%3A"+word, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	body, err := client.Get(req)
	if err != nil {
		return "", err
	}
	root, err := query.Parse(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	dict := root.Div(Class("lr_dct_ent"))
	def := ""
	if dictHTML := dict.Render(); dictHTML != nil {
		/*
			if audioLink := dict.Audio().Src(); audioLink != nil {
				audio = *audioLink
			}
		*/
		def = *dictHTML
	}
	if def == "" {
		err := errors.New("cannot find definition")
		// if page := root.Render(); page != nil {
		// err = errors.New(*page)
		// }
		return "", err

	}
	return def, nil
}

var rxSpace = regexp.MustCompile(`[\t \r\n!()\[\]\{\};:",<.>?“”‘’*/]+`)

func splitWords(s string) []string {
	words := rxSpace.Split(s, -1)
	m := make(map[string]bool)
	var results []string
	for _, word := range words {
		if word == "" {
			continue
		}
		if _, exists := m[word]; !exists {
			m[word] = true
			results = append(results, word)
		}
	}
	return words
}

type cache struct {
	m  map[string][]byte
	mu sync.RWMutex
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

	c.mu.RLock()
	if body, ok := c.m[uri]; ok {
		c.mu.RUnlock()
		return body, nil
	}
	c.mu.RUnlock()

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

	c.mu.Lock()
	c.m[uri] = body
	c.mu.Unlock()

	return body, nil
}

func prefetch(cache Getter, words []string) error {
	for _, word := range words {
		if _, err := googleDefinition(cache, word); err != nil {
			log.Print(err)
		}
		if _, err := googleImages(cache, word); err != nil {
			log.Print(err)
		}
		time.Sleep(time.Second)
	}
	return nil
}
