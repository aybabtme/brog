package brogger

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func SetUpDefaultBrog() *Brog {
	config := newDefaultConfig()
	config.ConsoleVerbosity = "info"
	logmux, _ := makeLogMux(config)
	b := &Brog{
		logMux: logmux,
		Config: config,
		isProd: false,
	}
	return b
}

func MakeBrogMultilingual(b *Brog) {
	b.Config.Multilingual = true
	b.Config.Languages = []string{"en", "fr"}
}

func TestExtractLanguage(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:3000/?en", nil)
	b := SetUpDefaultBrog()
	MakeBrogMultilingual(b)
	lang, _ := b.extractLanguage(req)
	if lang != "en" {
		t.Error("?en does not produce language of en")
	}
	req.URL, _ = url.Parse("http://localhost:3000/?fr")
	lang, _ = b.extractLanguage(req)
	if lang != "fr" {
		t.Error("?fr does not produce language of fr")
	}
	req.URL, _ = url.Parse("http://localhost:3000/")
	cookie := &http.Cookie{
		Name:  "lang",
		Value: "en",
	}
	req.AddCookie(cookie)
	lang, _ = b.extractLanguage(req)
	if lang != "en" {
		t.Error("lang=en cookie does not produce language of en")
	}
	cookie = &http.Cookie{
		Name:  "lang",
		Value: "fr",
	}
	req, _ = http.NewRequest("GET", "http://localhost:3000/", nil)
	_, langSet := b.extractLanguage(req)
	if langSet != false {
		t.Error("Language is set despite no GET parameters and no lang cookie.")
	}
	req.AddCookie(cookie)
	lang, _ = b.extractLanguage(req)
	if lang != "fr" {
		t.Error("lang=fr cookie does not produce language of fr")
	}
}

func TestSetLangCookie(t *testing.T) {
	b := SetUpDefaultBrog()
	MakeBrogMultilingual(b)
	req, _ := http.NewRequest("GET", "http://localhost:3000/?en", nil)
	rrw := httptest.NewRecorder()
	b.setLangCookie(req, rrw)
	if rrw.HeaderMap["Set-Cookie"][0] != "lang=en;Path=/" {
		t.Error("?en does not produce the right Set-Cookie header in the response")
	}
	cookie := &http.Cookie{
		Name:  "lang",
		Value: "en",
	}
	req.AddCookie(cookie)
	rrw = httptest.NewRecorder()
	b.setLangCookie(req, rrw)
	if rrw.HeaderMap["Set-Cookie"] != nil {
		t.Error("Set-Cookie header sent despite cookie being present when not referred from /changelang")
	}
	req.Header.Add("Referer", "http://localhost:3000/changelang")
	rrw = httptest.NewRecorder()
	b.setLangCookie(req, rrw)
	if rrw.HeaderMap["Set-Cookie"][0] != "lang=en;Path=/" {
		t.Error("Set-Cookie header not sent despite being referred from /changelang")
	}
}
