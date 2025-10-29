[ZAP LOGGING](https://github.com/uber-go/zap)

$env:ADDR=":8081"
$env:DB_ADDR="postgres://supervillager:adminpassword@localhost:5433/social?sslmode=disable"
$env:MIGRATIONS_PATH="./cmd/migrate/migrations"
$env:MAIL_TRAP_API_KEY="e9ae7e7015894ca627fb0a83ce47da15"
$env:REDIS_ENABLED="true"
$env:REDIS_ADDR="localhost:6379"
$env:RATE_LIMITER_ENABLED="true"
$env:RATELIMITER_REQUESTS_COUNT="5"


 siz `autocannon` bilan **load testing** qilayapsiz, ya’ni serveringizning **performance (TPS, RPS, latency, va stability)** ni tekshiryapsiz.


---

### ⚙️ Komanda tahlili:

```bash
npx autocannon -r 400 -d 2 -c 10 --renderStatusCodes http://localhost:8081/v1/health
```

#### Parametrlar:

| Parametr                          | Ma’nosi                                                                                                |
| --------------------------------- | ------------------------------------------------------------------------------------------------------ |
| `-r 400`                          | **Request rate limit** — soniyasiga 400 ta so‘rov yuborishga urinadi.                                  |
| `-d 2`                            | **Duration = 2 seconds** — test 2 soniya davom etadi.                                                  |
| `-c 10`                           | **Concurrency = 10 connections** — 10 ta parallel client bir vaqtda so‘rov yuboradi.                   |
| `--renderStatusCodes`             | Har bir **HTTP status code** (`200`, `404`, `500`, va hokazo) bo‘yicha nechtasi qaytganini ko‘rsatadi. |
| `http://localhost:8081/v1/health` | Siz test qilayotgan endpoint (bu yerda `health check`).                                                |

