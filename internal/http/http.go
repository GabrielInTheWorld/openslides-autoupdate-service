// Package http handles http requests for the autoupate service.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/autoupdate"
	"github.com/OpenSlides/openslides-autoupdate-service/internal/keysbuilder"
)

const prefix = "/system/autoupdate"

// Connecter returns an connect object.
type Connecter interface {
	Connect(userID int, kb autoupdate.KeysBuilder) autoupdate.DataProvider
}

// Autoupdate builds the requested keys from the body of a request. The
// body has to be in the format specified in the keysbuilder package.
func Autoupdate(mux *http.ServeMux, auth Authenticater, connecter Connecter) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		defer r.Body.Close()
		uid := auth.FromContext(r.Context())

		queryBuilder, err := keysbuilder.FromKeys(strings.Split(r.URL.Query().Get("k"), ","))
		if err != nil {
			handleError(w, fmt.Errorf("building keysbuilder from query: %w", err), true)
			return
		}

		bodyBuilder, err := keysbuilder.ManyFromJSON(r.Body)
		if err != nil {
			handleError(w, fmt.Errorf("building keysbuilder from body: %w", err), true)
			return
		}

		builder := keysbuilder.FromBuilders(queryBuilder, bodyBuilder)

		sender := sendMessages
		if r.URL.Query().Has("single") {
			sender = sendSingleMessage
		}

		if err := sender(r.Context(), w, uid, builder, connecter); err != nil {
			handleError(w, err, false)
			return
		}
	})

	mux.Handle(prefix, validRequest(authMiddleware(handler, auth)))
}

func sendMessages(ctx context.Context, w io.Writer, uid int, kb autoupdate.KeysBuilder, connecter Connecter) error {
	next := connecter.Connect(uid, kb)
	encoder := json.NewEncoder(w)

	for ctx.Err() == nil {
		// conn.Next() blocks, until there is new data. It also unblocks,
		// when the client context or the server is closed.
		data, err := next(ctx)
		if err != nil {
			return fmt.Errorf("getting next message: %w", err)
		}

		converted := make(map[string]json.RawMessage, len(data))
		for k, v := range data {
			converted[k] = v
		}

		if err := encoder.Encode(converted); err != nil {
			return fmt.Errorf("encoding and sending next message: %w", err)
		}

		w.(http.Flusher).Flush()
	}
	return ctx.Err()
}

func sendSingleMessage(ctx context.Context, w io.Writer, uid int, kb autoupdate.KeysBuilder, connecter Connecter) error {
	next := connecter.Connect(uid, kb)
	encoder := json.NewEncoder(w)

	// conn.Next() blocks, until there is new data. It also unblocks,
	// when the client context or the server is closed.
	data, err := next(ctx)
	if err != nil {
		return fmt.Errorf("getting next message: %w", err)
	}

	converted := make(map[string]json.RawMessage, len(data))
	for k, v := range data {
		converted[k] = v
	}

	if err := encoder.Encode(converted); err != nil {
		return fmt.Errorf("encoding end sending next message: %w", err)
	}
	return nil
}

// Health tells, if the service is running.
func Health(mux *http.ServeMux) {
	url := prefix + "/health"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		fmt.Fprintln(w, `{"healthy": true}`)
	})

	mux.Handle(url, handler)
}

func authMiddleware(next http.Handler, auth Authenticater) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := auth.Authenticate(w, r)
		if err != nil {
			handleError(w, fmt.Errorf("authenticate request: %w", err), true)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// handleError interprets the given error and writes a corresponding message to
// the client and/or stdout.
//
// If the handler already started to write the body then it is not allowed to
// set the http-status-code. In this case, writeStatusCode has to be fales.
func handleError(w http.ResponseWriter, err error, writeStatusCode bool) {
	if writeStatusCode {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	var closing interface {
		Closing()
	}
	if errors.As(err, &closing) {
		// Server is closing.
		return
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		// Client closed connection.
		return
	}

	var errClient ClientError
	if errors.As(err, &errClient) {
		if writeStatusCode {
			w.WriteHeader(http.StatusBadRequest)
		}

		fmt.Fprintf(w, `{"error": {"type": "%s", "msg": "%s"}}`, errClient.Type(), quote(errClient.Error()))
		return
	}

	if writeStatusCode {
		w.WriteHeader(http.StatusInternalServerError)
	}
	log.Printf("Internal Error: %v", err)
	fmt.Fprintln(w, `{"error": {"type": "InternalError", "msg": "Ups, something went wrong!"}}`)
}

// quote decodes changes quotation marks with a backslash to make sure, they are
// valid json.
func quote(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func validRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET or POST requests.
		if !(r.Method == http.MethodPost || r.Method == http.MethodGet) {
			handleError(w, invalidRequestError{fmt.Errorf("Only GET or POST requests are supported")}, true)
			return
		}

		next.ServeHTTP(w, r)
	})
}
