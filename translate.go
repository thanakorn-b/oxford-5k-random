package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type myMemoryResponse struct {
	ResponseData struct {
		TranslatedText string `json:"translatedText"`
	} `json:"responseData"`
	ResponseStatus int `json:"responseStatus"`
}

func translateEnToTh(ctx context.Context, client *http.Client, text, contactEmail string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", errors.New("empty text")
	}
	u, err := url.Parse("https://api.mymemory.translated.net/get")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("q", text)
	q.Set("langpair", "en|th")
	if contactEmail != "" {
		q.Set("de", contactEmail)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %s", resp.Status)
	}
	var out myMemoryResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if out.ResponseStatus != 200 {
		return "", fmt.Errorf("translation API status %d", out.ResponseStatus)
	}
	th := strings.TrimSpace(out.ResponseData.TranslatedText)
	if th == "" {
		return "", errors.New("empty translation")
	}
	return th, nil
}

const translationCacheVersion = 1

type cachedTranslation struct {
	Index int    `json:"index"`
	Th    string `json:"th"`
}

type translationCacheFile struct {
	Version int                          `json:"version"`
	ByWord  map[string]cachedTranslation `json:"by_word"`
}

type translationCache struct {
	path string
	data translationCacheFile
}

func loadTranslationCache(path string) (*translationCache, error) {
	c := &translationCache{
		path: path,
		data: translationCacheFile{
			Version: translationCacheVersion,
			ByWord:  make(map[string]cachedTranslation),
		},
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, &c.data); err != nil {
		return nil, err
	}
	if c.data.ByWord == nil {
		c.data.ByWord = make(map[string]cachedTranslation)
	}
	c.data.Version = translationCacheVersion
	return c, nil
}

func (c *translationCache) save() error {
	c.data.Version = translationCacheVersion
	raw, err := json.MarshalIndent(&c.data, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	dir := filepath.Dir(c.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.path)
}

func (c *translationCache) lookup(word string) (string, bool) {
	if c == nil || c.data.ByWord == nil {
		return "", false
	}
	v, ok := c.data.ByWord[word]
	if !ok || strings.TrimSpace(v.Th) == "" {
		return "", false
	}
	return v.Th, true
}

func (c *translationCache) put(word string, index int, th string) error {
	if c == nil {
		return nil
	}
	if c.data.ByWord == nil {
		c.data.ByWord = make(map[string]cachedTranslation)
	}
	c.data.ByWord[word] = cachedTranslation{Index: index, Th: th}
	return c.save()
}

func translateWithCache(ctx context.Context, client *http.Client, email string, cache *translationCache, word string, index int) (string, error) {
	word = strings.TrimSpace(word)
	if word == "" {
		return "", errors.New("empty text")
	}
	if cache != nil {
		if th, ok := cache.lookup(word); ok {
			return th, nil
		}
	}
	th, err := translateEnToTh(ctx, client, word, email)
	if err != nil {
		return "", err
	}
	if cache != nil {
		if err := cache.put(word, index, th); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save translation cache: %v\n", err)
		}
	}
	return th, nil
}
