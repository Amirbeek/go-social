package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/amirbeek/social/internal/auth"
	"github.com/amirbeek/social/internal/db"
	"github.com/amirbeek/social/internal/env"
	"github.com/amirbeek/social/internal/mailer"
	"github.com/amirbeek/social/internal/ratelimiter"
	store2 "github.com/amirbeek/social/internal/store"
	"github.com/amirbeek/social/internal/store/cache"
	cache2 "github.com/amirbeek/social/internal/store/cache"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"

	"go.uber.org/zap"
)

const version string = "v0.1"

//	@title			GopherSocial
//	@description	API for GopherSocial, a social network for gahpers
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8081
//	@BasePath	/v1

//	@securityDefinitions.basic	BasicAuth

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {
	// Load configuration from environment
	cfg := config{
		Addr:        env.GetString("ADDR", ":8081"), // default port :8081
		apiURL:      env.GetString("EXTERNAL_URL", "localhost:8081"),
		frontendURL: env.GetString("FRONTEND_URL", "localhost:8081"),
		DB: dbConfig{
			Addr:         env.GetString("DB_ADDR", "postgres://supervillager:adminpassword@localhost:5433/social?sslmode=disable"),
			MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 5),
			MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
			MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		redisConfig: redisConfig{
			addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
			db:      env.GetInt("REDIS_DB", 0),
			pw:      env.GetString("REDIS_PASSWORD", ""),
			enabled: env.GetBool("REDIS_ENABLED", false),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "hello@amirbekshomurodov.me"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey: env.GetString("MAIL_TRAP_API_KEY", ""),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("AUTH_BASIC_USER", "admin"),
				pass: env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "gophersocial",
			},
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATELIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Log rate limiter config early
	logger.Infow("Rate limiter config",
		"enabled", cfg.rateLimiter.Enabled,
		"limit", cfg.rateLimiter.RequestsPerTimeFrame,
		"window", cfg.rateLimiter.TimeFrame.String(),
	)

	// Connect to database
	dbConn, err := db.New(
		cfg.DB.Addr,
		cfg.DB.MaxOpenConns,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxIdleTime)
	if err != nil {
		logger.Fatal("Database connection failed: %v", err)
	}

	defer dbConn.Close()
	logger.Info("Database connection pool established")

	// Cache
	var rdb *redis.Client
	if cfg.redisConfig.enabled {
		rdb = cache2.NewRedisClient(cfg.redisConfig.addr, cfg.redisConfig.pw, cfg.redisConfig.db) // <-- FIXED (rdb = not :=)
		logger.Info("Redis connection pool established")
	}

	// Mail server
	//mailer := mailer.NewSendGrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apiKey, cfg.mail.fromEmail)
	if err != nil {
		logger.Fatal("Mailtrap client failed: %v", err)
	}

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	// Rate Limiter
	rateLimiter := ratelimiter.NewFixedWindowLimiter(
		cfg.rateLimiter.RequestsPerTimeFrame,
		cfg.rateLimiter.TimeFrame,
	)

	// storage layer
	store := store2.NewStorage(dbConn)
	cacheStorage := cache.NewRedisStorage(rdb)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: jwtAuthenticator,
		rateLimiter:   rateLimiter,
	}

	// Log rate limiter status before starting server
	if cfg.rateLimiter.Enabled {
		logger.Infow("Rate limiter ENABLED",
			"limit", cfg.rateLimiter.RequestsPerTimeFrame,
			"window", cfg.rateLimiter.TimeFrame.String(),
		)
	} else {
		logger.Info("Rate limiter DISABLED")
	}

	mux := app.mount()

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return dbConn.Stats()
	}))

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	logger.Infof("Server starting at http://localhost%s...", cfg.Addr)

	if err := app.run(mux); err != nil {
		logger.Fatalf("Server stopped with error: %v", err)
	}
}
