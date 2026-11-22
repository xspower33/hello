package main

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

//go:embed web
var webFS embed.FS

type pageData struct {
	Greeting string
	Date     string
	Time     string
}

func main() {
	loc := loadTZ()
	now := func() time.Time { return time.Now().In(loc) }

	tmpl := template.Must(template.ParseFS(webFS, "web/index.html"))
	static, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		t := now()
		data := pageData{
			Greeting: greeting(t.Hour()),
			Date:     formatDate(t),
			Time:     t.Format("15:04"),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			log.Printf("render: %v", err)
		}
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           nosniff(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("hello listening on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func nosniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

// loadTZ returns the zone named by $TZ, falling back to the system local zone.
func loadTZ() *time.Location {
	if name := os.Getenv("TZ"); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
		log.Printf("unknown TZ %q, using system local time", name)
	}
	return time.Local
}

func greeting(hour int) string {
	switch {
	case hour >= 5 && hour < 12:
		return "günaydın"
	case hour >= 12 && hour < 18:
		return "iyi günler"
	case hour >= 18 && hour < 23:
		return "iyi akşamlar"
	default:
		return "iyi geceler"
	}
}

var trMonths = [...]string{
	"Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
	"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
}

var trDays = [...]string{
	"Pazar", "Pazartesi", "Salı", "Çarşamba", "Perşembe", "Cuma", "Cumartesi",
}

func formatDate(t time.Time) string {
	return strconv.Itoa(t.Day()) + " " + trMonths[t.Month()-1] + " " +
		strconv.Itoa(t.Year()) + ", " + trDays[t.Weekday()]
}
