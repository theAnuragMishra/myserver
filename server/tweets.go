package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TweetModule(pool *pgxpool.Pool) RouterFunc {
	return func(r chi.Router) {
		r.Route("/tweets", func(r chi.Router) {
			r.Get("/get", getTweetsHandler(pool))
			r.Group(func(r chi.Router) {
				r.Use(adminVerificationMiddleware)
				r.Post("/post", postTweetHandler(pool))
			})
		})
	}
}

func getTweetsHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		rows, err := pool.Query(ctx, `
            SELECT id, text, created_at 
            FROM tweets 
            ORDER BY created_at DESC
        `)
		if err != nil {
			http.Error(w, "failed to fetch tweets", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type Tweet struct {
			ID        int       `json:"id"`
			Text      string    `json:"text"`
			CreatedAt time.Time `json:"created_at"`
		}

		var tweets []Tweet

		for rows.Next() {
			var t Tweet
			err := rows.Scan(&t.ID, &t.Text, &t.CreatedAt)
			if err != nil {
				http.Error(w, "failed to read tweets", http.StatusInternalServerError)
				return
			}
			tweets = append(tweets, t)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "error iterating tweets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tweets)
	}
}

func postTweetHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var tweet string
		err := json.NewDecoder(r.Body).Decode(&tweet)
		if err != nil {
			http.Error(w, "Invalid tweet content", http.StatusBadRequest)
			return
		}

		_, err = pool.Exec(r.Context(), "INSERT INTO tweets(text, created_at) VALUES ($1, $2)", tweet, time.Now())
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}
}
