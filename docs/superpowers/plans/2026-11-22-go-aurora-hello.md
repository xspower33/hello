# hello — Go Aurora Karşılama Ekranı: Uygulama Planı

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Go standart kütüphanesiyle servis edilen, aurora arka planlı buzlu cam karşılama ekranı olan ve `docker compose up` ile tek komutla çalışan bir web uygulaması.

**Architecture:** Tek statik Go binary. HTML/CSS/JS `embed.FS` ile binary'ye gömülür, `html/template` ile sunucu tarafında ilk selamlama basılır, `app.js` saati canlı tutar. Çok aşamalı Dockerfile ile `alpine` üzerinde root olmayan kullanıcı olarak çalışır; compose tek servis tanımlar.

**Tech Stack:** Go 1.24 (stdlib: `net/http`, `html/template`, `embed`, `log/slog`), saf CSS/JS, Docker (multi-stage), Docker Compose.

**Spec:** `docs/superpowers/specs/2026-11-22-go-aurora-hello-design.md`

## Global Constraints

- Yalnızca Go standart kütüphanesi; `go.mod` içinde hiçbir `require` satırı olmayacak.
- Modül adı `hello`, Go sürümü `1.24`.
- Harici font/CDN/JS kütüphanesi yok. Tipografi: `system-ui, -apple-system, "Segoe UI", Roboto, sans-serif`.
- Selamlama eşikleri: 05–11 `günaydın`, 12–17 `iyi günler`, 18–22 `iyi akşamlar`, 23–04 `iyi geceler`.
- Tarih biçimi: `22 Kasım 2026, Pazar` (gün adı baş harfi büyük).
- Varsayılan port `8080` (`PORT` env), varsayılan saat dilimi `Europe/Istanbul` (`TZ` env).
- Arka plan rengi `#080b14`; cam kart `max-width: 620px`, `backdrop-filter: blur(24px) saturate(140%)`.
- `prefers-reduced-motion: reduce` altında hiçbir animasyon çalışmayacak.
- Konteyner root olmayan kullanıcı (uid `10001`) ile çalışacak.
- Docker imajı 50 MB altında kalacak.

## Review Focus

- **Saat dilimi:** Konteyner UTC'de çalışırsa 21:00'de "iyi günler" görünür. `TZ=Europe/Istanbul` verilmediğinde sunucu kendi varsayılanına düşmeli, çökmemeli. (Task 2)
- **Gece yarısı geçişi:** Saat 23:59 → 00:00 olduğunda hem tarih hem selamlama güncellenmeli; sayfa açık kalırken bayat kalmamalı. (Task 3)
- **Kök dışı yollar:** `/yok` gibi bir yol 404 dönmeli, index'e düşmemeli. (Task 2)
- **Sağlık kontrolü:** `/healthz` sadece süreç ayaktayken 200 dönmeli ve `wget` ile kontrol edilebilir olmalı. (Task 2, Task 4)
- **Hareket hassasiyeti:** `prefers-reduced-motion` açıkken aurora ve giriş animasyonu durmalı, içerik görünür kalmalı. (Task 3)

---

### Task 1: Modül ve saat yardımcıları

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Test: `main_test.go`

**Interfaces:**
- Consumes: yok
- Produces: `greeting(hour int) string`, `formatDate(t time.Time) string`, `loadTZ() *time.Location`

- [ ] **Step 1: Modülü oluştur**

`go.mod` içeriği:

```
module hello

go 1.24
```

- [ ] **Step 2: Başarısız testleri yaz**

`main_test.go`:

```go
package main

import (
	"testing"
	"time"
)

func TestGreeting(t *testing.T) {
	cases := map[int]string{
		0: "iyi geceler", 4: "iyi geceler",
		5: "günaydın", 11: "günaydın",
		12: "iyi günler", 17: "iyi günler",
		18: "iyi akşamlar", 22: "iyi akşamlar",
		23: "iyi geceler",
	}
	for hour, want := range cases {
		if got := greeting(hour); got != want {
			t.Errorf("greeting(%d) = %q, want %q", hour, got, want)
		}
	}
}

func TestFormatDate(t *testing.T) {
	loc := time.FixedZone("TRT", 3*60*60)
	got := formatDate(time.Date(2026, time.November, 22, 21, 41, 0, 0, loc))
	if want := "22 Kasım 2026, Pazar"; got != want {
		t.Errorf("formatDate() = %q, want %q", got, want)
	}
}

func TestLoadTZ(t *testing.T) {
	t.Setenv("TZ", "Europe/Istanbul")
	if got := loadTZ().String(); got != "Europe/Istanbul" {
		t.Errorf("loadTZ() = %q, want Europe/Istanbul", got)
	}
	t.Setenv("TZ", "Mars/Olympus")
	if got := loadTZ(); got == nil {
		t.Fatal("loadTZ() returned nil for an unknown zone")
	}
}
```

- [ ] **Step 3: Testlerin başarısız olduğunu doğrula**

Run: `go test ./...`
Expected: FAIL — `undefined: greeting`, `undefined: formatDate`, `undefined: loadTZ`

- [ ] **Step 4: Yardımcıları yaz**

`main.go` — `greeting`, `formatDate` (Türkçe ay/gün tabloları ile) ve `loadTZ` (`os.Getenv("TZ")` → `time.LoadLocation`, hata veya boşsa `time.Local`). `greeting` için `switch` ile aralık kontrolü kullan.

- [ ] **Step 5: Testlerin geçtiğini doğrula**

Run: `go test ./... -run 'TestGreeting|TestFormatDate|TestLoadTZ' -v`
Expected: PASS, 3 test

- [ ] **Step 6: Commit**

```bash
git add go.mod main.go main_test.go
git commit -m "feat: add greeting and date helpers"
```

---

### Task 2: HTTP sunucusu, route'lar ve gömülü dosyalar

**Files:**
- Modify: `main.go`
- Create: `web/index.html` (geçici iskelet), `web/style.css`, `web/app.js` (boş dosyalar)
- Test: `main_test.go`

**Interfaces:**
- Consumes: Task 1'den `greeting`, `formatDate`, `loadTZ`
- Produces: `type pageData struct { Greeting, Date, Time string }`, `newHandler(tmpl *template.Template, static fs.FS, now func() time.Time) http.Handler`

- [ ] **Step 1: Geçici web dosyalarını oluştur**

`web/index.html` en az `<html><body>{{.Greeting}}</body></html>` içersin; `web/style.css` ve `web/app.js` boş olabilir. (Go'nun `//go:embed web` derlemesi için dosyaların var olması gerekir. Task 3 bunları gerçek içerikle değiştirecek.)

- [ ] **Step 2: Başarısız testleri yaz**

`main_test.go` dosyasına ekle:

```go
func testHandler(t *testing.T) http.Handler {
	t.Helper()
	tmpl := template.Must(template.ParseFS(webFS, "web/index.html"))
	static, err := fs.Sub(webFS, "web")
	if err != nil {
		t.Fatal(err)
	}
	fixed := time.Date(2026, time.November, 22, 21, 41, 7, 0, time.FixedZone("TRT", 3*60*60))
	return newHandler(tmpl, static, func() time.Time { return fixed })
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, httptest.NewRequest("GET", "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}
	if got, want := rec.Body.String(), `{"status":"ok"}`; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestIndexRendersGreetingAndDate(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"iyi akşamlar", "22 Kasım 2026, Pazar", "21:41"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, httptest.NewRequest("GET", "/yok", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestStaticFile(t *testing.T) {
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, httptest.NewRequest("GET", "/static/style.css", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/css") {
		t.Errorf("Content-Type = %q, want text/css", got)
	}
}
```

`import` bloğuna `html/template`, `io/fs`, `net/http`, `net/http/httptest`, `strings` ekle.

- [ ] **Step 3: Testlerin başarısız olduğunu doğrula**

Run: `go test ./... -run 'TestHealthz|TestIndex|TestNotFound|TestStatic'`
Expected: FAIL — `undefined: webFS`, `undefined: newHandler`

- [ ] **Step 4: Sunucuyu yaz**

`main.go`'ya ekle:

- `//go:embed web` + `var webFS embed.FS`
- `pageData` struct'ı (`Greeting`, `Date`, `Time`)
- `newHandler(tmpl *template.Template, static fs.FS, now func() time.Time) http.Handler`:
  - `mux := http.NewServeMux()`
  - `GET /{$}` → `tmpl.ExecuteTemplate(w, "index.html", pageData{...})`; saat `now().Format("15:04")`; hata olursa `slog.Error` + `http.Error(w, ..., 500)`
  - `GET /healthz` → `Content-Type: application/json`, gövde tam olarak `{"status":"ok"}`
  - `GET /static/` → `http.StripPrefix("/static/", http.FileServer(http.FS(static)))`
  - Tüm yanıtlara `X-Content-Type-Options: nosniff` ekleyen küçük bir middleware sarmala
- `main()`:
  - `loadTZ()` sonucunu `time.Local` yerine kullanmak için `now := func() time.Time { return time.Now().In(loc) }`
  - `template.Must(template.ParseFS(webFS, "web/index.html"))`
  - `PORT` env (boşsa `8080`), `:PORT` adresinde `http.Server{ReadHeaderTimeout: 5 * time.Second}`
  - `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` + `srv.Shutdown` (10 sn timeout)

- [ ] **Step 5: Testlerin geçtiğini doğrula**

Run: `go test ./... -v`
Expected: PASS, tüm testler

- [ ] **Step 6: Sunucuyu elle dene**

Run: `go run .` (ayrı terminalde `curl -s localhost:8080/healthz`)
Expected: `{"status":"ok"}`

- [ ] **Step 7: Commit**

```bash
git add main.go main_test.go web/
git commit -m "feat: serve embedded index page with health endpoint"
```

---

### Task 3: Aurora karşılama ekranı

**Files:**
- Modify: `web/index.html`, `web/style.css`, `web/app.js`

**Interfaces:**
- Consumes: `pageData` alanları — şablonda `{{.Greeting}}`, `{{.Date}}`, `{{.Time}}`
- Produces: DOM kimlikleri `#clock`, `#date`, `#greeting` (Task 3'ün JS'i bunları günceller)

- [ ] **Step 1: HTML iskeletini yaz**

`web/index.html`: `lang="tr"`, `<title>hello — karşılama</title>`, `/static/style.css` ve `/static/app.js` referansları. Gövde:

```
<main class="stage">
  <section class="card">
    <p class="label">HELLO</p>
    <h1 id="clock" aria-hidden="true">{{.Time}}</h1>
    <p class="sr-only">Şu anki saat: {{.Time}}</p>
    <p id="date">{{.Date}}</p>
    <p id="greeting">{{.Greeting}}</p>
    <p class="hint">docker compose up · Go ile servis edildi</p>
  </section>
</main>
<div class="aurora" aria-hidden="true">
  <span class="blob b1"></span><span class="blob b2"></span>
  <span class="blob b3"></span><span class="blob b4"></span>
</div>
```

Aurora katmanı `.stage`'in arkasında kalmalı (`z-index` ile).

- [ ] **Step 2: Aurora ve cam stillerini yaz**

`web/style.css`:

- `:root` değişkenleri: `--bg: #080b14`, `--ink: #eef2ff`, `--muted: #9aa7c7`, palet renkleri (menekşe `#7c5cff`, camgöbeği `#22d3ee`, pembe `#f472b6`, amber `#fbbf24`)
- `body` → `background: var(--bg)`, sistem font yığını, `min-height: 100dvh`, `overflow: hidden`
- `.aurora` → `position: fixed; inset: 0; z-index: 0; filter: blur(80px); opacity: .55`
- `.blob` → `position: absolute; width: 45vmax; aspect-ratio: 1; border-radius: 50%; background: radial-gradient(circle, <renk> 0%, transparent 70%)`
- Dört blob farklı `top/left` ve `animation: drift-* 26s/34s/41s/29s ease-in-out infinite alternate` ile sürüklensin; keyframe'ler `transform: translate3d/scale` kullanır (layout tetiklemez)
- `.stage` → `position: relative; z-index: 1; display: grid; place-items: center; min-height: 100dvh; padding: 24px`
- `.card` → `max-width: 620px; width: 100%; padding: 48px 40px; border-radius: 24px; text-align: center; background: rgba(255,255,255,.06); border: 1px solid rgba(255,255,255,.12); backdrop-filter: blur(24px) saturate(140%); box-shadow: 0 24px 80px rgba(0,0,0,.45); animation: rise .6s ease-out both`
- `#clock` → `font-size: clamp(3rem, 12vw, 5.5rem); font-variant-numeric: tabular-nums; letter-spacing: -.03em; font-weight: 600`
- `.label` → `letter-spacing: .35em; font-size: .75rem; color: var(--muted)` ve altında 1px ince çizgi
- `.hint` → `font-size: .8rem; color: var(--muted); opacity: .8`
- `.sr-only` → ekrandan gizleme standardı
- `@keyframes rise` → `from { opacity: 0; transform: translateY(16px) } to { opacity: 1; transform: none }`
- `@media (prefers-reduced-motion: reduce)` → `*, *::before, *::after { animation: none !important; transition: none !important }`

- [ ] **Step 3: Canlı saati yaz**

`web/app.js`:

- `greetingFor(hour)` — Task 1'deki eşiklerin aynısı
- `pad(n)` yardımcısı
- `render()` → `new Date()` ile `HH:MM:SS` üret, `#clock` ve `#date`'i `Intl.DateTimeFormat("tr-TR", {day:"numeric", month:"long", year:"numeric", weekday:"long"})` ile güncelle, `#greeting`'i `greetingFor(d.getHours())` ile güncelle
- `setInterval(render, 1000)` + ilk çağrı
- Tarih biçimini `Intl` çıktısına birebir güvenmek yerine `day/month/year` parçalarından kur ki `22 Kasım 2026, Pazar` biçimi korunsun (gün adı büyük harfle başlasın)

- [ ] **Step 4: İşaretleri doğrula**

Run: `grep -c "blob" web/style.css` ve `grep -n "prefers-reduced-motion" web/style.css`
Expected: `blob` en az 5 kez, `prefers-reduced-motion` bulunur

Run: `grep -n "setInterval" web/app.js`
Expected: bir eşleşme

- [ ] **Step 5: Sunucu üzerinden kontrol et**

Run: `go run .` → tarayıcıda `http://localhost:8080`
Expected: koyu zemin, hareketli aurora, ortada cam kart, canlı saat, `iyi akşamlar` (saate göre)

- [ ] **Step 6: Commit**

```bash
git add web/
git commit -m "feat: add aurora glass welcome screen"
```

---

### Task 4: Docker, compose ve README

**Files:**
- Create: `Dockerfile`, `docker-compose.yml`, `.dockerignore`, `.gitignore`, `README.md`

**Interfaces:**
- Consumes: Task 2'nin `PORT`/`TZ` env sözleşmesi ve `/healthz` endpoint'i
- Produces: `docker compose up --build` ile `http://localhost:8080` üzerinde çalışan servis

- [ ] **Step 1: Dockerfile yaz**

`Dockerfile`:

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hello .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/hello /app/hello
USER app
ENV PORT=8080 TZ=Europe/Istanbul
EXPOSE 8080
ENTRYPOINT ["/app/hello"]
```

- [ ] **Step 2: Compose dosyasını yaz**

`docker-compose.yml` — `version` anahtarı yok:

```yaml
services:
  web:
    build: .
    image: hello:latest
    ports:
      - "${APP_PORT:-8080}:8080"
    environment:
      PORT: "8080"
      TZ: "Europe/Istanbul"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

- [ ] **Step 3: Yok sayma dosyalarını yaz**

`.dockerignore`: `.git`, `docs`, `README.md`, `*.md`, `docker-compose.yml`
`.gitignore`: `/hello`, `/hello.exe`, `.DS_Store`

- [ ] **Step 4: README yaz**

`README.md` (Türkçe): ne olduğu bir paragraf, `docker compose up --build` ve `http://localhost:8080`, `APP_PORT` ile port değiştirme, `go run .` ile Docker'sız çalıştırma, `TZ` notu, dosya yapısı tablosu.

- [ ] **Step 5: Yapılandırmayı doğrula**

Run: `docker compose config`
Expected: hata yok, `web` servisi ve `healthcheck` görünür

- [ ] **Step 6: Commit**

```bash
git add Dockerfile docker-compose.yml .dockerignore .gitignore README.md
git commit -m "build: add Dockerfile and compose setup"
```

---

### Task 5: Uçtan uca doğrulama

**Files:**
- Modify: yok (gerekirse hata düzeltmeleri)

**Interfaces:**
- Consumes: Task 1–4'ün tamamı
- Produces: çalıştığı kanıtlanmış proje

- [ ] **Step 1: Testleri çalıştır**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 2: İmajı derle ve ayağa kaldır**

Run: `docker compose up --build -d`
Expected: imaj derlenir, servis başlar

- [ ] **Step 3: Sağlık ve içerik kontrolü**

Run: `curl -s http://localhost:8080/healthz`
Expected: `{"status":"ok"}`

Run: `curl -s http://localhost:8080/ | grep -o "iyi [a-z]*"`
Expected: saate uygun selamlama

- [ ] **Step 4: Sağlık durumunu kontrol et**

Run: `docker compose ps`
Expected: `healthy`

- [ ] **Step 5: İmaj boyutunu kontrol et**

Run: `docker image inspect hello:latest --format '{{.Size}}'`
Expected: `52428800` (50 MB) altında

- [ ] **Step 6: Kapat**

Run: `docker compose down`

- [ ] **Step 7: Commit (düzeltme yapıldıysa)**

```bash
git add -A
git commit -m "fix: address issues found in end-to-end verification"
```
