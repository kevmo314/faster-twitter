package types

import (
	"bytes"
	"net/url"
	"strings"
	"text/template"
	"time"

	"golang.org/x/exp/slices"
)

// Tweet represents a Twitter Tweet, previously called a status.
// https://dev.twitter.com/overview/api/tweets
type Tweet struct {
	ID            string    `json:"id"`
	Text          string    `json:"text"`
	CreatedAt     time.Time `json:"created_at"`
	AuthorID      string    `json:"author_id"`
	PublicMetrics struct {
		RetweetCount int `json:"retweet_count"`
		ReplyCount   int `json:"reply_count"`
		LikeCount    int `json:"like_count"`
		QuoteCount   int `json:"quote_count"`
	} `json:"public_metrics"`
	Entities struct {
		URLs []struct {
			Start       int    `json:"start"`
			End         int    `json:"end"`
			URL         string `json:"url"`
			ExpandedURL string `json:"expanded_url"`
			DisplayURL  string `json:"display_url"`
			Images      []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"images"`
			Status      int    `json:"status"`
			Title       string `json:"title"`
			Description string `json:"description"`
			UnwoundURL  string `json:"unwound_url"`
			MediaKey    string `json:"media_key"`
		} `json:"urls"`
		Hashtags []struct {
			Start int    `json:"start"`
			End   int    `json:"end"`
			Tag   string `json:"tag"`
		} `json:"hashtags"`
		Mentions []struct {
			Start    int    `json:"start"`
			End      int    `json:"end"`
			Username string `json:"username"`
		} `json:"mentions"`
		Cashtags []struct {
			Start int    `json:"start"`
			End   int    `json:"end"`
			Tag   string `json:"tag"`
		} `json:"cashtags"`
	} `json:"entities"`
	Attachments struct {
		MediaKeys []string `json:"media_keys"`
	} `json:"attachments"`
}

func (t Tweet) RenderHTML() string {
	type Replacement struct {
		start, end int
		html       []rune
	}

	imageTmpl, err := template.New("tweet").Parse(`<a class="main-url-link" href="{{.URL.URL}}">
		<div class="main-url-link-image">
			<img src="{{.Image.URL}}" width="{{.Image.Width}}" height="{{.Image.Height}}">
		</div>
		<div class="main-url-link-host">
			{{.Host}}
		</div>
		<div class="main-url-link-title">
			{{.URL.Title}}
		</div>
		<div class="main-url-link-description hide-sm">
			{{.URL.Description}}
		</div>
	</a>`)
	if err != nil {
		panic(err)
	}

	replacements := []Replacement{}
	for _, u := range t.Entities.URLs {
		replacements = append(replacements, Replacement{u.Start, u.End, []rune(`<a href="` + u.URL + `">` + u.DisplayURL + `</a>`)})
	}
	for _, hashtag := range t.Entities.Hashtags {
		replacements = append(replacements, Replacement{hashtag.Start, hashtag.End, []rune(`<a href="https://twitter.com/hashtag/` + hashtag.Tag + `">#` + hashtag.Tag + `</a>`)})
	}
	for _, mention := range t.Entities.Mentions {
		replacements = append(replacements, Replacement{mention.Start, mention.End, []rune(`<a href="https://twitter.com/` + mention.Username + `">@` + mention.Username + `</a>`)})
	}
	for _, cashtag := range t.Entities.Cashtags {
		replacements = append(replacements, Replacement{cashtag.Start, cashtag.End, []rune(`<a href="https://twitter.com/search?q=%24` + cashtag.Tag + `">$` + cashtag.Tag + `</a>`)})
	}
	// sort replacements by start index
	slices.SortFunc(replacements, func(a, b Replacement) bool {
		return a.start < b.start
	})
	html := []rune(t.Text)
	// apply replacements in reverse
	for i := len(replacements) - 1; i >= 0; i-- {
		r := replacements[i]
		html = append(html[:r.start], append(r.html, html[r.end:]...)...)
	}
	for _, u := range t.Entities.URLs {
		if u.Images != nil {
			parsed, err := url.Parse(u.ExpandedURL)
			if err != nil {
				panic(err)
			}
			var mainURLHTML bytes.Buffer
			imageTmpl.Execute(&mainURLHTML, map[string]interface{}{
				"URL":   u,
				"Host":  strings.TrimPrefix(parsed.Host, "www."),
				"Image": u.Images[0],
			})
			html = append(html, []rune(mainURLHTML.String())...)
		}
	}
	return string(html)
}

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
}

type Media struct {
	Type       string `json:"type"`
	MediaKey   string `json:"media_key"`
	DurationMs int    `json:"duration_ms"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Variants   []struct {
		ContentType string `json:"content_type"`
		URL         string `json:"url"`
	} `json:"variants"`
}

type PublicMetrics struct {
	RetweetCount int `json:"retweet_count"`
	ReplyCount   int `json:"reply_count"`
	LikeCount    int `json:"like_count"`
	QuoteCount   int `json:"quote_count"`
}
