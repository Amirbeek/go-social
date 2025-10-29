package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal error",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusInternalServerError, "the server encountered an error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("bad request",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusBadRequest, err.Error())
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	writeJsonError(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("conflict response",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusConflict, err.Error())
}
func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("forbidden response",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusForbidden, "forbidden")
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("not found error",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusNotFound, "not found")
}

func (app *application) unauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized error",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw("unauthorized basic error",
		"method", r.Method,
		"path", r.URL.Path,
		"err", err,
	)
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted", charset="UTF-8"`)
	writeJsonError(w, http.StatusUnauthorized, "unauthorized")
}
