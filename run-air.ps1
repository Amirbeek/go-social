$env:ADDR = ":8081"
$env:DB_ADDR = "postgres://supervillager:adminpassword@localhost:5433/social?sslmode=disable"
$env:MIGRATIONS_PATH = "./cmd/migrate/migrations"
$env:MAIL_TRAP_API_KEY = "e9ae7e7015894ca627fb0a83ce47da15"

air
