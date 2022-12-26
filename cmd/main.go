package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kevmo314/faster-twitter/pkg/types"
)

const twitterAPIKey = "kNYoLSsjCQrvRNBSNyNR7UV5L"
const twitterAPISecret = "SdeKRX5kBq1UBeHneENPr8YNirA8BshLNy4LimSyn8QOIHUk5c"
const twitterBearerToken = "AAAAAAAAAAAAAAAAAAAAAMNFkwEAAAAAkebFYhjEepEY9703pvIR2zKM7%2BY%3DRiK008cIuQgTxbzBMK9zpliixVbTUuECdQcCWtmDx7pChenBZM"

func tzOffset(r *http.Request) time.Duration {
	tzOffsetStr := r.URL.Query().Get("tz_offset")
	tzOffset, err := time.ParseDuration(tzOffsetStr)
	if err != nil {
		return 0
	}
	return tzOffset
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		fn(gzr, r)
	}
}
func main() {
	http.HandleFunc("/embed/", makeGzipHandler(func(w http.ResponseWriter, r *http.Request) {
		// the tweet id is the last part of the path
		tweetID := r.URL.Path[len("/embed/"):]
		// fetch the tweet from the twitter API
		req := &http.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme: "https",
				Host:   "api.twitter.com",
				Path:   "/2/tweets/" + tweetID,
				RawQuery: url.Values{
					"expansions":   []string{"author_id"},
					"tweet.fields": []string{"created_at,public_metrics,entities"},
					"user.fields":  []string{"name,username,profile_image_url"},
				}.Encode(),
			},
			Header: http.Header{
				"Authorization": []string{"Bearer " + twitterBearerToken},
			},
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// decode as a tweet
		var tweet struct {
			Data     types.Tweet `json:"data"`
			Includes struct {
				Users []types.User `json:"users"`
			} `json:"includes"`
		}
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&tweet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// render the tweet
		tmpl, err := template.ParseFiles("web/embed.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data struct {
			Tweet     types.Tweet
			Author    types.User
			Time      string
			Date      string
			Permalink string
			HTML      template.HTML
		}

		data.Tweet = tweet.Data
		data.Author = tweet.Includes.Users[0]
		data.Time = data.Tweet.CreatedAt.Add(tzOffset(r)).Format("3:04 PM")
		data.Date = data.Tweet.CreatedAt.Add(tzOffset(r)).Format("Jan 2, 2006")
		data.Permalink = "https://twitter.com/" + data.Author.Username + "/status/" + data.Tweet.ID
		data.HTML = template.HTML(data.Tweet.RenderHTML())

		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("error executing template: %v", err)
			return
		}
	}))
	http.HandleFunc("/widgets.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/widgets.js")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
