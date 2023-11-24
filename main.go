package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/nathanielfernandes/cnvs/preview"

	_ "embed"
)

// canvas api (quilt)
var canvasBaseUrl = mustGetEnvString("CANVAS_BASE_URL")
var canvasSecret = mustGetEnvString("CANVAS_SECRET")

//go:embed preview.ql
var previewGenCode string

var CACHE = make(map[string]*bytes.Buffer)
var LOCK = sync.RWMutex{}

func main() {
	go cleanCacheLoop()

	// token.StartAccessTokenReferesher()
	preview.StartScrapeRunner()

	router := httprouter.New()
	router.GET("/:trackID", track)
	router.GET("/:trackID/info", track_info)
	router.GET("/:trackID/audio", redirect_to_audio)

	fmt.Println("Listening on port 80")
	if err := http.ListenAndServe("0.0.0.0:80", router); err != nil {
		log.Fatal(err)
	}
}

func cleanCacheLoop() {
	for range time.Tick(time.Hour) {
		LOCK.Lock()
		CACHE = make(map[string]*bytes.Buffer)
		LOCK.Unlock()
	}
}

func track(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackID := ps.ByName("trackID")

	buf, err := getOrGen(trackID)
	if err != nil {
		http.Error(w, "Failed to generate track", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Write(buf.Bytes())
}

func track_info(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackID := ps.ByName("trackID")

	pre, err := preview.GetPreview(trackID)
	if err != nil {
		http.Error(w, "Failed to get preview", http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(pre)
	if err != nil {
		http.Error(w, "Failed to marshal preview", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func redirect_to_audio(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	trackID := ps.ByName("trackID")

	pre, err := preview.GetPreview(trackID)
	if err != nil {
		http.Error(w, "Failed to get preview", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, pre.AudioURL, http.StatusFound)
}

func mustGetEnvString(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return val
}

func getOrGen(trackID string) (*bytes.Buffer, error) {
	LOCK.RLock()
	buf, ok := CACHE[trackID]
	LOCK.RUnlock()

	if ok {
		return buf, nil
	}

	pre, err := preview.GetPreview(trackID)
	if err != nil {
		return nil, err
	}

	// weird bug where indexing is out of bounds
	artistName := "unkown"
	// join multiple artists with commas
	artistNames := []string{}
	for _, artist := range pre.Artists {
		if artist.Name != "" {
			artistNames = append(artistNames, artist.Name)
		}
	}

	if len(artistNames) > 0 {
		artistName = strings.Join(artistNames, ", ")
	}

	buf, err = generateImage(pre.CoverArt.Small, pre.TrackName, artistName, pre.BackgroundColor)
	if err != nil {
		return nil, err
	}

	LOCK.Lock()
	CACHE[trackID] = buf
	LOCK.Unlock()

	return buf, nil
}

func generateImage(album_art, track_name, artist_name, bg_color string) (*bytes.Buffer, error) {
	payload := RunPayload{
		Size: []int{512, 670},
		Files: []File{
			{
				Name: "preview.ql",
				Code: previewGenCode,
			},
		},
		Assets: []interface{}{
			ImageAsset{
				Name: "art",
				Url:  album_art,
			},
			LiteralAsset{
				Name:    "track_name",
				Literal: fmt.Sprintf("\"%s\"", track_name),
			},
			LiteralAsset{
				Name:    "artist_name",
				Literal: fmt.Sprintf("\"%s\"", artist_name),
			},
			LiteralAsset{
				Name:    "color",
				Literal: bg_color,
			},
			LiteralAsset{
				Name:    "spotify_logo",
				Literal: "\"M248 8C111.1 8 0 119.1 0 256s111.1 248 248 248 248-111.1 248-248S384.9 8 248 8zm100.7 364.9c-4.2 0-6.8-1.3-10.7-3.6-62.4-37.6-135-39.2-206.7-24.5-3.9 1-9 2.6-11.9 2.6-9.7 0-15.8-7.7-15.8-15.8 0-10.3 6.1-15.2 13.6-16.8 81.9-18.1 165.6-16.5 237 26.2 6.1 3.9 9.7 7.4 9.7 16.5s-7.1 15.4-15.2 15.4zm26.9-65.6c-5.2 0-8.7-2.3-12.3-4.2-62.5-37-155.7-51.9-238.6-29.4-4.8 1.3-7.4 2.6-11.9 2.6-10.7 0-19.4-8.7-19.4-19.4s5.2-17.8 15.5-20.7c27.8-7.8 56.2-13.6 97.8-13.6 64.9 0 127.6 16.1 177 45.5 8.1 4.8 11.3 11 11.3 19.7-.1 10.8-8.5 19.5-19.4 19.5zm31-76.2c-5.2 0-8.4-1.3-12.9-3.9-71.2-42.5-198.5-52.7-280.9-29.7-3.6 1-8.1 2.6-12.9 2.6-13.2 0-23.3-10.3-23.3-23.6 0-13.6 8.4-21.3 17.4-23.9 35.2-10.3 74.6-15.2 117.5-15.2 73 0 149.5 15.2 205.4 47.8 7.8 4.5 12.9 10.7 12.9 22.6 0 13.6-11 23.3-23.2 23.3z\"",
			},
		},
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(jsonBytes)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", canvasBaseUrl+"/run/"+canvasSecret, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("canvas api returned %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	return buf, nil
}

type RunPayload struct {
	Size   []int         `json:"size"`
	Files  []File        `json:"files"`
	Assets []interface{} `json:"assets"`
}

type File struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type ImageAsset struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type LiteralAsset struct {
	Name    string `json:"name"`
	Literal string `json:"literal"`
}
