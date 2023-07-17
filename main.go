package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/nathanielfernandes/cnvs/preview"
	"github.com/nathanielfernandes/cnvs/token"

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

	token.StartAccessTokenReferesher()
	preview.StartPreviewRunner()

	router := httprouter.New()
	router.GET("/:trackID", track)

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

	buf, err = generateImage(pre.CoverArtURL, pre.TrackName, pre.ArtistName)
	if err != nil {
		return nil, err
	}

	LOCK.Lock()
	CACHE[trackID] = buf
	LOCK.Unlock()

	return buf, nil
}

func generateImage(album_art, track_name, artist_name string) (*bytes.Buffer, error) {
	payload := RunPayload{
		Size: []int{512, 640},
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
