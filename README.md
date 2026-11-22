# hello

Go ile yazılmış iki minik servis. Tek `docker compose` komutuyla ayağa kalkar.

| Servis | Port (host) | Port (konteyner) | Ne yapar |
| --- | --- | --- | --- |
| `web` | `80` | `4380` | Aurora arka planlı karşılama ekranı: canlı saat, saate göre selamlama |
| `notfound` | `4381` | `4381` | Animasyonlu **404** sayfası + rastgele veri uç noktaları |

## Çalıştırma

```bash
docker compose up --build
```

- Karşılama ekranı: <http://localhost>
- 404 servisi: <http://localhost:4381>

Portları değiştirmek için:

```bash
APP_PORT=9000 docker compose up --build          # macOS/Linux
$env:APP_PORT=9000; docker compose up             # PowerShell
```

Aynı şekilde `NOTFOUND_PORT` ile 404 servisinin host portunu değiştirebilirsin.

> **Not:** Varsayılan host portu `80`. Windows'ta bu portu başka bir servis (IIS, Skype, VS Code port forwarding) kullanıyorsa ya da izin hatası alırsan `APP_PORT` ile başka bir porta al ya da Docker Desktop'ı yönetici olarak çalıştır.

Kapatmak için: `docker compose down`

## Docker olmadan

Go 1.22+ kuruluysa:

```bash
go run .                      # karşılama ekranı, 4380
go run ./notfound             # 404 servisi, 4381
```

## Dokploy ile yayınlama

Depo iki ayrı Dokploy uygulaması olarak yayınlanır; ikisi de **Dockerfile** derleyicisini kullanır.

**Uygulama 1 — karşılama ekranı**

| Ayar | Değer |
| --- | --- |
| Build Type | `Dockerfile` |
| Dockerfile Path | `Dockerfile` |
| Port | `4380` |

**Uygulama 2 — 404 servisi**

| Ayar | Değer |
| --- | --- |
| Build Type | `Dockerfile` |
| Dockerfile Path | `notfound/Dockerfile` |
| Port | `4381` |

Ortam değişkeni olarak istersen `TZ=Europe/Istanbul` ekle (imajlarda zaten varsayılan).
Nixpacks seçili kaldığı sürece repodaki `Dockerfile`'lar kullanılmaz.

Konteynerler root olmayan `app` (uid 10001) kullanıcısıyla çalışır, imajlar ~20 MB'tır.

### Alt alan adı önerisi

Sunucunun IP'sini bilen ücretsiz bir host adı yeterli; `sslip.io` host içindeki IP'ye çözümlenir:

| Uygulama | Örnek host |
| --- | --- |
| karşılama | `hello.<SUNUCU-IP>.sslip.io` |
| 404 | `404.<SUNUCU-IP>.sslip.io` |

Host adında **alt çizgi kullanma** (yalnızca harf, rakam, tire ve nokta).

## Uç noktalar

### Karşılama servisi (4380)

| Yol | Açıklama |
| --- | --- |
| `/` | Karşılama ekranı |
| `/healthz` | `{"status":"ok"}` |
| `/static/...` | CSS ve JS |

### 404 servisi (4381)

`/` ve `/healthz`, `/random/*` ve `/static/*` dışındaki **her yol** animasyonlu 404 sayfasını döner.

| Yol | Açıklama |
| --- | --- |
| `/` | 404 tasarımının önizlemesi (200) |
| `/healthz` | `{"status":"ok"}` |
| `/random/number` | Rastgele tam sayı — `?min=&max=` |
| `/random/color` | Rastgele renk — `{"hex":"#7c5cff","rgb":[124,92,255]}` |
| `/random/quote` | Rastgele bir söz |
| `/random/coin` | Yazı tura |
| `/random/dice` | Zar atışı — `?faces=` |
| `/random/uuid` | Rastgele UUID v4 |
| `/random/endpoint` | Rastgele bir uç noktayı tarif eder |
| diğer | Animasyonlu 404 sayfası |

Örnek:

```bash
curl -s localhost:4381/random/color
curl -s "localhost:4381/random/number?min=1&max=6"
curl -s localhost:4381/random/endpoint
```

404 sayfasındaki uç nokta etiketlerine tıklayarak yanıtı sayfanın içinde görebilirsin;
"rastgele uç nokta" düğmesi `/random/endpoint`'ten bir ad seçip onu çağırır.

## Ayarlar

| Değişken | Varsayılan | Açıklama |
| --- | --- | --- |
| `APP_PORT` | `80` | Karşılama servisinin host portu (konteyner içi hep `4380`) |
| `NOTFOUND_PORT` | `4381` | 404 servisinin host portu (konteyner içi `4381`) |
| `PORT` | `4380` / `4381` | Konteyner içinde dinlenen port |
| `TZ` | `Europe/Istanbul` | Saat dilimi (tzdata binary'ye gömülüdür) |

## Dosyalar

| Yol | İçerik |
| --- | --- |
| `main.go` | Karşılama servisi: route'lar, saat ve selamlama yardımcıları |
| `web/index.html` | Karşılama sayfası iskeleti (Go şablonu) |
| `web/style.css` | Aurora arka plan ve cam kart |
| `web/app.js` | Canlı saat ve selamlama |
| `Dockerfile` | Karşılama servisi imajı |
| `notfound/main.go` | 404 servisi: 404 tasarımı ve rastgele uç noktalar |
| `notfound/web/404.html` | 404 sayfası iskeleti (Go şablonu) |
| `notfound/web/notfound.css` | Radar, ızgara, parçacık ve glitch animasyonları |
| `notfound/web/notfound.js` | Uç nokta çağrıları ve yanıt paneli |
| `notfound/Dockerfile` | 404 servisi imajı |
| `docker-compose.yml` | İki servis: `web` ve `notfound` |
