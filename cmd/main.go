package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]string{"status": "ok"}
	_ = json.NewEncoder(w).Encode(resp)
}

func listUsers(w http.ResponseWriter, _ *http.Request) {
	type user struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	const (
		fetchUsersUrl = "https://jsonplaceholder.typicode.com/users"
	)

	var (
		client       = http.Client{Timeout: time.Second * 3}
		request, _   = http.NewRequest("GET", fetchUsersUrl, nil)
		fetchedUsers []user
	)

	resp, err := client.Do(request)
	if err != nil {
		handleError(w, err)
		return
	}

	defer func() { _ = resp.Body.Close() }()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&fetchedUsers); err != nil {
		handleError(w, err)
		return
	}

	_ = json.NewEncoder(w).Encode(fetchedUsers)
}

func handleError(w http.ResponseWriter, err error) {
	resp := map[string]string{"err": err.Error()}
	_ = json.NewEncoder(w).Encode(resp)
}

func respContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func doGracefulShutDown() {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
		done        = make(chan struct{})
	)

	defer cancel()

	go doCloseResources(ctx, done)

	for {
		select {
		case <-ctx.Done():
			log.Printf("closing resources has timed out")
			return
		case <-done:
			log.Printf("closing resources has finished on time")
			return
		}
	}
}

func doCloseResources(_ context.Context, done chan<- struct{}) {
	log.Print("closing resources ....")
	time.Sleep(time.Second * 3)
	log.Print("done closing resources ....")
	done <- struct{}{}
}

func serveHttp() {
	const (
		httpServingPort = ":7000"
	)

	var (
		sigCh  = make(chan os.Signal)
		router = mux.NewRouter().StrictSlash(true)
	)

	signal.Notify(sigCh, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		log.Printf("received sig: %v", sig)
		doGracefulShutDown()
		os.Exit(0)
	}()

	router.HandleFunc("/", healthCheck).Methods("GET")
	router.HandleFunc("/users", listUsers).Methods("GET")
	router.Use(respContentTypeMiddleware)

	log.Printf("serving app at: %s", httpServingPort)
	log.Fatal(http.ListenAndServe(httpServingPort, router))
}

func main() {
	serveHttp()
}
