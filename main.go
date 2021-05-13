package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-pg/pg/v10/orm"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	maxShutdownTime = 10 * time.Second

	countFakeUsers = 10000
)

var (
	port = flag.Uint("port", 8000, "server port")
	dsn  = flag.String("dsn", "postgresql://app:Hid8ujik@app-postgresql/app?sslmode=disable", "postgres dsn")
)

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"OK\"}"))
}

func root(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ROOT HANDLER] request: %v %v", r.Method, r.URL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("root handler\n"))
}

func newServer(address string, svc *Service) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz/", health)
	mux.HandleFunc("/", root)
	mux.HandleFunc("/user/count", svc.GetCountUsers)
	mux.HandleFunc("/user/random", svc.GetRandomUser)
	mux.Handle("/metrics", promhttp.Handler())
	return &http.Server{Addr: address, Handler: mux}
}

func main() {
	rand.Seed(time.Now().Unix())

	flag.Parse()
	address := fmt.Sprintf(":%d", *port)

	svc, err := NewService(*dsn)
	if err != nil {
		log.Printf("create service: %v", err)
		return
	}
	srv := newServer(address, svc)

	shutdown := make(chan struct{})
	defer close(shutdown)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		sig := <-stop
		log.Printf("Got signal '%v', the graceful shutdown will start", sig)

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, maxShutdownTime)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		} else {
			log.Print("HTTP server has been shutdown")
		}
		shutdown <- struct{}{}
	}()

	log.Printf("Start HTTP server %s", address)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	log.Print("Wait for the shutdown server")
	<-shutdown
}

// Service

type Service struct {
	db *pg.DB
}

func (s *Service) GetCountUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	if err := s.db.Model(&users).Select(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("{count of user: %d}", len(users))))
}

func (s *Service) GetRandomUser(w http.ResponseWriter, r *http.Request) {
	id := rand.Int63n(countFakeUsers)
	userName := fmt.Sprintf("fake-%d", id)
	log.Printf("get user %s", userName)
	user := User{}
	err := s.db.Model(&user).
		Where("name = ?", userName).
		Limit(1).
		Select()
	if err != nil {
		log.Printf("get random user (id: %d): %v", id, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	body, err := json.Marshal(user)
	if err != nil {
		log.Printf("marshal user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func NewService(dsn string) (*Service, error) {

	opt, err := pg.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	db := pg.Connect(opt)
	err = createSchema(db)
	if err != nil {
		log.Printf("create schema: %v", err)
	} else {
		log.Printf("create schema: done")
	}
	err = createFakeUsers(db, countFakeUsers)
	if err != nil {
		log.Printf("create fake users: %v", err)
	}
	return &Service{db: db}, nil
}

// Database

type User struct {
	Id     int64
	Name   string
	Emails []string
}

func (u User) String() string {
	return fmt.Sprintf("User<%d %s %v>", u.Id, u.Name, u.Emails)
}

func createSchema(db *pg.DB) error {
	models := []interface{}{
		(*User)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func createFakeUsers(db *pg.DB, count int) error {
	err := db.Model(&User{}).Where("name = ?", "fake-0").Select()
	if err == nil {
		log.Printf("users already exists")
		return nil
	}
	users := make([]User, count)
	for i := 0; i < count; i++ {
		users[i] = User{
			Name:   fmt.Sprintf("fake-%d", i),
			Emails: []string{fmt.Sprintf("fake-master-%d@email.com", i), fmt.Sprintf("fake-slave-%d@email.com", i)},
		}
	}

	_, err = db.Model(&users).Insert()
	return err
}
