package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"local/complex-web/server/database"
	"local/complex-web/server/queue"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq" // To register the driver.
)

func main() {
	mainCtx := context.Background()

	dbService, err := database.GetConnectionPostgres()
	if err != nil {
		slog.Error("Failed to connect to database", "err", err)
		os.Exit(1)
	}

	queueService, err := queue.InitRedisService(mainCtx)
	if err != nil {
		slog.Error("Failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer func() {
		err := queueService.Close()
		if err != nil {
			slog.Error("Failed to close redis", "err", err)
			os.Exit(1)
		}
	}()

	if dbService == nil || queueService == nil {
		slog.Error("Failed to init dependencies")
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /add", func(writer http.ResponseWriter, request *http.Request) {
		slog.Debug("Received POST /add")
		err := request.ParseForm()
		if err != nil {
			slog.Error("Failed to parse form", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))

			return
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			slog.Error("Failed to read body", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}
		defer request.Body.Close()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			slog.Error("Failed to parse body", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}

		var (
			num any
			ok  bool
		)
		if num, ok = data["num"]; !ok {
			slog.Error("Failed to parse body", "err", "num is missing")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}

		err = queueService.Write(request.Context(), num.(float64))
		if err != nil {
			slog.Error("Failed to write", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(fmt.Sprintf(`Added value %v`, num)))

	})

	mux.HandleFunc("GET /get/{num}", func(writer http.ResponseWriter, request *http.Request) {
		slog.Debug("Handle GET /get/{num}")
		num, err := strconv.Atoi(request.PathValue("num"))
		if err != nil {
			slog.Error("Failed to parse value", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}

		res, err := dbService.Get(request.Context(), num)
		if err != nil {
			slog.Error("Failed to get value", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		jsonBytes, err := json.Marshal(res)
		if err != nil {
			slog.Error("Failed to marshal value", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Header().Add("Content-Type", "application/json")
		_, err = writer.Write(jsonBytes)
		if err != nil {
			slog.Error("Failed to marshal value", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			writer.Write([]byte(err.Error()))
			return
		}
		slog.Debug("Sending back GET /get/{num} -> " + string(jsonBytes))
		return
	})

	log.Println("Listening on :3001")
	log.Fatal(http.ListenAndServe(":3001", mux))
}
