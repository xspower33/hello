# hello

Go ile yazılmış minik bir karşılama ekranı. Tek statik binary, tek `docker compose` komutu.

Aurora arka plan, ortada buzlu cam kart, canlı saat, saate göre selamlama
(`günaydın` / `iyi günler` / `iyi akşamlar` / `iyi geceler`).

## Çalıştırma

```bash
docker compose up --build
```

Sonra tarayıcıda: <http://localhost:8080>

Portu değiştirmek için:

```bash
APP_PORT=9000 docker compose up --build     # macOS/Linux
$env:APP_PORT=9000; docker compose up        # PowerShell
```

Kapatmak için: `docker compose down`

## Docker olmadan

Go 1.24+ kuruluysa:

```bash
go run .
```

## Uç noktalar

| Yol | Açıklama |
| --- | --- |
| `/` | Karşılama ekranı |
| `/healthz` | `{"status":"ok"}` |
| `/static/...` | CSS ve JS |

## Ayarlar

| Değişken | Varsayılan | Açıklama |
| --- | --- | --- |
| `PORT` | `8080` | Dinlenen port |
| `TZ` | `Europe/Istanbul` | Saat ve selamlama için saat dilimi |

## Dosyalar

| Yol | İçerik |
| --- | --- |
| `main.go` | Sunucu, route'lar, saat ve selamlama yardımcıları |
| `web/index.html` | Sayfa iskeleti (Go şablonu) |
| `web/style.css` | Aurora arka plan ve cam kart |
| `web/app.js` | Canlı saat ve selamlama |
| `Dockerfile` | Çok aşamalı derleme, root olmayan kullanıcı |
| `docker-compose.yml` | Tek servis `web`, healthcheck |
