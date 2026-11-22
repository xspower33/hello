# hello — Go karşılama ekranı (Aurora / cam) — Tasarım

Tarih: 2026-11-22
Durum: onaylandı (yön: B — Aurora / cam)

## Niyet

`xspower33/hello` deposu boştu. Amaç: Go ile yazılmış, `docker compose up` ile tek komutla
ayağa kalkan, `index` sayfasında hoş bir karşılama ekranı olan küçük ama derli toplu bir
proje. "Kafana bir şey yap" denildiği için görsel yön üç alternatif arasından seçildi.

Neden: deponun adı `hello`; ilk izlenim olarak sıcak, modern, "buraya hoş geldin" diyen bir
ekran istendi.

## Hedefler

- Tek komutla çalışsın: `docker compose up`
- Bağımlılıksız Go sunucusu (yalnızca standart kütüphane)
- Karşılama ekranı: canlı saat, tarih, saate göre selamlama, aurora arka plan, buzlu cam kart
- Küçük imaj (~20 MB), root olmayan kullanıcı, healthcheck
- JavaScript kapalıyken de anlamlı içerik (sunucu tarafı selamlama)
- Mobilde düzgün, `prefers-reduced-motion` destekli

## Hedef olmayanlar

- Kimlik doğrulama, veritabanı, API, çoklu sayfa
- Harici font/CDN/JS kütüphanesi (imaj çevrimdışı da çalışsın)
- Kubernetes/Traefik/HTTPS sonlandırma

## Anti-hedefler

- Şişkin framework (gin/echo), node build adımı, `node_modules`
- Sadece "hello world" yazan ruhsuz bir sayfa
- Docker imajının 100 MB üstüne çıkması

## Kısıtlar

- Geliştirme makinesinde `go` ve `docker` kurulu değil → doğrulama kullanıcı Docker Desktop
  kurduktan sonra yapılabilir. Kod buna göre yazılır, komutlar README'de verilir.
- Uzak depo boş; `main` dalında ilk commit bizden gidecek.

## Yaklaşımlar

1. **Saf stdlib + gömülü statik dosyalar** (seçilen) — `embed` ile HTML/CSS/JS binary'ye gömülür,
   tek statik binary. Basit, hızlı, bağımlılık yok.
2. Framework (chi/gin) + harici şablonlar — route yönetimi kolay ama tek sayfa için gereksiz
   bağımlılık ve daha büyük imaj.
3. Node/React ön yüz + Go API — bu projenin kapsamı için fazla; build zinciri ve imaj boyutu
   artar.

Seçim: **1**. Tek sayfalık bir karşılama ekranı için en az hareketli parça.

## Tasarım

### Dosya yapısı

```
hello/
├── main.go              # sunucu, route'lar, graceful shutdown
├── go.mod
├── Dockerfile           # çok aşamalı: golang:1.24-alpine → alpine
├── docker-compose.yml   # tek servis "web"
├── .dockerignore
├── .gitignore
├── README.md
└── web/
    ├── index.html       # html/template, sunucu tarafı selamlama
    ├── style.css        # aurora + cam
    └── app.js           # canlı saat + selamlama güncellemesi
```

### Sunucu (`main.go`)

- `net/http` + `html/template` + `embed.FS`; framework yok
- Route'lar:
  - `GET /` → `index.html` şablonu (selamlama, tarih sunucuda hesaplanır)
  - `GET /static/` → gömülü `web/` dizini (`http.FileServer` + `fs.Sub`)
  - `GET /healthz` → `{"status":"ok"}`, `application/json`
  - Diğer yollar → 404
- Ayarlar: `PORT` (varsayılan `8080`), `TZ` (varsayılan `Europe/Istanbul`)
  - Konteynerler varsayılan olarak UTC çalışır; bu yüzden saat ve selamlama yanlış olurdu.
    `TZ` verilmezse sunucu `Europe/Istanbul` yüklemeyi dener, başarısız olursa `time.Local`.
- `signal.NotifyContext` ile graceful shutdown, 10 sn timeout
- Sadece okuma amaçlı olduğu için statik dosyalarda `Cache-Control: no-cache` (geliştirme
  kolaylığı), güvenlik başlıkları: `X-Content-Type-Options: nosniff`

### Sayfa (görsel)

- Tam ekran koyu zemin (`#080b14` üzerine hafif dikey gradyan)
- **Aurora**: 4 adet büyük, bulanık (`filter: blur(80px)`) radyal gradyan leke; farklı
  sürelerde yavaşça sürüklenir/yaylanır (CSS keyframes, `transform` + `opacity`). Üstte ince
  bir grain katmanı (`repeating-radial-gradient`, düşük opaklık) ile bant izleri kırılır.
- **Cam kart**: ortada, `max-width: 620px`, `backdrop-filter: blur(24px) saturate(140%)`,
  `background: rgba(255,255,255,0.06)`, 1px yarı saydam kenar, yumuşak gölge, 24px köşe
- Kart içeriği:
  - Üstte küçük, aralıklı harfli `HELLO` etiketi + ince çizgi
  - Büyük saat `22:41:07` (tabular rakamlar, saniye canlı)
  - Altında tarih: `22 Kasım 2026, Pazar`
  - Selamlama: `iyi akşamlar` (saate göre: 05–11 günaydın, 12–17 iyi günler,
    18–22 iyi akşamlar, 23–04 iyi geceler)
  - Altta küçük ipucu satırı: `docker compose up · Go ile servis edildi`
- Hareket: kart girişte 600 ms `translateY + opacity` ile belirir; aurora sürekli ama yavaş
- `prefers-reduced-motion: reduce` → tüm animasyonlar durur, kart statik görünür
- Tipografi: sistem font yığını (`system-ui, -apple-system, "Segoe UI", ...`), harici font yok
- Erişilebilirlik: tek `h1`, kontrast AA, saat `aria-hidden="true"` + görünmez metin özeti,
  semantik `main`

### JavaScript (`app.js`)

- Saniyede bir saati ve dakikada bir selamlamayı günceller (`Intl.DateTimeFormat("tr-TR")`)
- JS yoksa sunucunun bastığı değerler kalır (kademeli geliştirme)

### Docker

- `Dockerfile`:
  - Aşama 1 `golang:1.24-alpine`: `CGO_ENABLED=0 GOOS=linux` ile statik derleme,
    `-trimpath -ldflags="-s -w"`
  - Aşama 2 `alpine:3.20`: `ca-certificates` yok (dış çağrı yok), `adduser -D -u 10001 app`,
    binary `/app/hello`, `EXPOSE 8080`, `USER app`, `ENTRYPOINT ["/app/hello"]`
  - Healthcheck için alpine'daki busybox `wget` kullanılır
- `docker-compose.yml`: `version` anahtarı yok (güncel şema), servis `web`,
  `ports: "${APP_PORT:-8080}:8080"`, `environment: PORT=8080` ve `TZ=Europe/Istanbul`,
  `restart: unless-stopped`, healthcheck (`wget -qO- http://127.0.0.1:8080/healthz`,
  30 sn aralık, 3 deneme, 5 sn start_period)

## Doğrulama

Docker kurulu bir makinede:

1. `docker compose up --build -d`
2. `curl http://localhost:8080/healthz` → `{"status":"ok"}`
3. Tarayıcıda `http://localhost:8080` → aurora arka plan, cam kart, canlı saat
4. `docker compose ps` → servis `healthy`
5. `docker compose down`

Docker yoksa `go run .` ile de çalışır (Go kuruluysa).

## Yapıcıya kalanlar

- Aurora lekelerinin tam renk paleti ve süreleri (görsel zevk)
- Grain katmanının yoğunluğu
- README'nin Türkçe/İngilizce dili
