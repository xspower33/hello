package main

import (
	"context"
	crand "crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	_ "time/tzdata"
)

//go:embed web
var webFS embed.FS

// endpoint, sayfada listelenen ve /random/endpoint tarafından seçilen bir uç nokta.
type endpoint struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

var endpoints = []endpoint{
	{Path: "/random/number", Description: "rastgele tam sayı (?min=&max=)"},
	{Path: "/random/color", Description: "rastgele renk"},
	{Path: "/random/quote", Description: "rastgele bir söz"},
	{Path: "/random/coin", Description: "yazı tura"},
	{Path: "/random/dice", Description: "zar atışı (?faces=)"},
	{Path: "/random/uuid", Description: "rastgele uuid v4"},
	{Path: "/random/endpoint", Description: "rastgele bir uç nokta"},
}

var quotes = []struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}{
	{"Yol, yürüyenin ayağı altında uzar.", "Anonim"},
	{"404, kaybın değil aramanın kanıtıdır.", "Bu sayfa"},
	{"Kaybolmak, başka bir yere varmaktır.", "Anonim"},
	{"Aradığın şey burada değil; belki hiç olmadı.", "Sunucu"},
	{"Her yanlış yol, doğru yolun bir elenmesidir.", "Anonim"},
	{"Haritada olmayan yerler de vardır.", "Anonim"},
	{"Bir sayfa bulunamazsa, tarayıcı yine de güzeldir.", "Anonim"},
}

type pageData struct {
	Path      string
	Method    string
	Endpoints []endpoint
}

func main() {
	tmpl := template.Must(template.ParseFS(webFS, "web/404.html"))
	static, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /random/number", handleNumber)
	mux.HandleFunc("GET /random/color", handleColor)
	mux.HandleFunc("GET /random/quote", handleQuote)
	mux.HandleFunc("GET /random/coin", handleCoin)
	mux.HandleFunc("GET /random/dice", handleDice)
	mux.HandleFunc("GET /random/uuid", handleUUID)
	mux.HandleFunc("GET /random/endpoint", handleEndpoint)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))

	// Kayıtlı olmayan her yol: 404 tasarımı. Kök yolu tasarımı görmek için 200 döner.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderNotFound(w, tmpl, r, r.URL.Path == "/")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "4381"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           nosniff(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("notfound listening on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// renderNotFound 404 tasarımını yazar. preview true ise durum kodu 200 bırakılır
// (kök yolu tasarımı görmek için kullanılır).
func renderNotFound(w http.ResponseWriter, tmpl *template.Template, r *http.Request, preview bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !preview {
		w.WriteHeader(http.StatusNotFound)
	}

	data := pageData{
		Path:      r.URL.Path,
		Method:    r.Method,
		Endpoints: endpoints,
	}
	if err := tmpl.ExecuteTemplate(w, "404.html", data); err != nil {
		log.Printf("render: %v", err)
	}
}

func nosniff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json: %v", err)
	}
}

func handleNumber(w http.ResponseWriter, r *http.Request) {
	min, max := 1, 100
	if v, err := strconv.Atoi(r.URL.Query().Get("min")); err == nil {
		min = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("max")); err == nil {
		max = v
	}
	if min > max {
		min, max = max, min
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"value": rand.IntN(max-min+1) + min,
		"min":   min,
		"max":   max,
	})
}

func handleColor(w http.ResponseWriter, r *http.Request) {
	code := randomColor()
	writeJSON(w, http.StatusOK, map[string]any{
		"hex": code,
		"rgb": []int{hexVal(code[1:3]), hexVal(code[3:5]), hexVal(code[5:7])},
	})
}

func handleQuote(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, quotes[rand.IntN(len(quotes))])
}

func handleCoin(w http.ResponseWriter, r *http.Request) {
	result := "yazı"
	if rand.IntN(2) == 1 {
		result = "tura"
	}
	writeJSON(w, http.StatusOK, map[string]string{"result": result})
}

func handleDice(w http.ResponseWriter, r *http.Request) {
	faces := 6
	if v, err := strconv.Atoi(r.URL.Query().Get("faces")); err == nil && v >= 2 {
		faces = v
	}
	writeJSON(w, http.StatusOK, map[string]int{"value": rand.IntN(faces) + 1, "faces": faces})
}

func handleUUID(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"uuid": randomUUID()})
}

func handleEndpoint(w http.ResponseWriter, r *http.Request) {
	// Kendini seçmesin: geri kalanlardan birini döndür.
	pool := endpoints[:len(endpoints)-1]
	writeJSON(w, http.StatusOK, pool[rand.IntN(len(pool))])
}

func randomColor() string {
	return fmt.Sprintf("#%02x%02x%02x", rand.IntN(256), rand.IntN(256), rand.IntN(256))
}

func hexVal(s string) int {
	v, _ := strconv.ParseInt(s, 16, 32)
	return int(v)
}

func randomUUID() string {
	var b [16]byte
	if _, err := crand.Read(b[:]); err != nil {
		// pratikte olmaz; yine de sayfayı düşürmeyelim
		for i := range b {
			b[i] = byte(rand.IntN(256))
		}
	}
	b[6] = (b[6] & 0x0f) | 0x40 // sürüm 4
	b[8] = (b[8] & 0x3f) | 0x80 // varyant
	s := hex.EncodeToString(b[:])
	return s[0:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:32]
}
