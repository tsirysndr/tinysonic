package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tsirysndr/tinysonic/internal/config"
	"github.com/tsirysndr/tinysonic/internal/scanner"
)

// State is shared by all handlers.
type State struct {
	Pool      *sql.DB
	Username  string
	Password  string
	MusicDir  string
	CoversDir string
	Progress  *scanner.Progress
}

func Start(ctx context.Context, cfg *config.Config, pool *sql.DB, progress *scanner.Progress) error {
	s := &State{
		Pool:      pool,
		Username:  cfg.Username,
		Password:  cfg.Password,
		MusicDir:  cfg.MusicDir,
		CoversDir: cfg.CoversDir,
		Progress:  progress,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)

	register := func(name string, h http.HandlerFunc) {
		mux.HandleFunc("/rest/"+name, h)
		mux.HandleFunc("/rest/"+name+".view", h)
	}

	// System
	register("ping", s.Ping)
	register("getUser", s.GetUser)
	register("getMusicFolders", s.GetMusicFolders)
	register("getScanStatus", s.GetScanStatus)
	register("startScan", s.StartScan)

	// Library — tag browsing
	register("getArtists", s.GetArtists)
	register("getArtist", s.GetArtist)
	register("getAlbum", s.GetAlbum)
	register("getSong", s.GetSong)
	register("getAlbumList2", s.GetAlbumList2)
	register("getAlbumList", s.GetAlbumList)

	// Folder browsing
	register("getIndexes", s.GetIndexes)
	register("getMusicDirectory", s.GetMusicDirectory)

	// Genres
	register("getGenres", s.GetGenres)
	register("getSongsByGenre", s.GetSongsByGenre)

	// Lists
	register("getRandomSongs", s.GetRandomSongs)
	register("getStarred", s.GetStarred)
	register("getStarred2", s.GetStarred2)

	// Playback
	register("stream", s.Stream)
	register("download", s.Stream)
	register("getCoverArt", s.GetCoverArt)
	register("scrobble", s.Scrobble)
	register("getNowPlaying", s.GetNowPlaying)
	register("updateNowPlaying", s.UpdateNowPlaying)

	// Search
	register("search3", s.Search3)
	register("search2", s.Search2)

	// Playlists
	register("getPlaylists", s.GetPlaylists)
	register("getPlaylist", s.GetPlaylist)
	register("createPlaylist", s.CreatePlaylist)
	register("updatePlaylist", s.UpdatePlaylist)
	register("deletePlaylist", s.DeletePlaylist)

	// Starring
	register("star", s.Star)
	register("unstar", s.Unstar)

	// Artist / album info stubs
	register("getArtistInfo", s.GetArtistInfo)
	register("getArtistInfo2", s.GetArtistInfo2)
	register("getAlbumInfo", s.GetAlbumInfo)
	register("getAlbumInfo2", s.GetAlbumInfo)
	register("getSimilarSongs", s.GetSimilarSongs)
	register("getSimilarSongs2", s.GetSimilarSongs2)
	register("getTopSongs", s.GetTopSongs)
	register("getLyrics", s.GetLyrics)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           corsMiddleware(mux),
		ReadHeaderTimeout: 15 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// corsMiddleware mirrors actix-cors permissive: allow any origin, any method,
// any header, with credentials.
func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if reqHdrs := r.Header.Get("Access-Control-Request-Headers"); reqHdrs != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHdrs)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "*")
		}
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// requireAuth returns true when the request authenticated; it has already
// written the error response when this returns false.
func (s *State) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	q := r.URL.Query()
	if !authCheck(s.Username, s.Password, q.Get("u"), q.Get("p"), q.Get("t"), q.Get("s")) {
		writeError(w, 40, "Wrong username or password")
		return false
	}
	return true
}

// queryFirst is a convenience for endpoints that take many params on the
// query string — both GET and POST are routed identically and the Subsonic
// spec puts params in the query string regardless of method.
func queryFirst(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

// queryAll returns every value for the given key, in the order they appeared.
func queryAll(r *http.Request, key string) []string {
	return r.URL.Query()[key]
}
