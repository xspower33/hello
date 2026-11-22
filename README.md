# hello

Go ile yazılmış minik bir karşılama ekranı. Tek statik binary, tek `docker compose` komutu.

Aurora arka plan, ortada buzlu cam kart, canlı saat, saate göre selamlama
(`günaydın` / `iyi günler` / `iyi akşamlar` / `iyi geceler`).

## Çalıştırma

```bash
docker compose up --build
```

Sonra tarayıcıda: <http://localhost>

Portu değiştirmek için:

```bash
APP_PORT=9000 docker compose up --build     # macOS/Linux
$env:APP_PORT=9000; docker compose up        # PowerShell
```

> **Not:** Varsayılan host portu `80`. Windows'ta bu portu başka bir servis (IIS, Skype, VS Code port forwarding) kullanıyorsa ya da izin hatası alırsan `APP_PORT` ile başka bir porta al ya da Docker Desktop'ı yönetici olarak çalıştır.

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
| `APP_PORT` | `80` | Host tarafında dışa açılan port (konteyner içi hep `8080`) |
| `PORT` | `8080` | Konteyner içinde dinlenen port |
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
