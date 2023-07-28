package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

const (
	// environment variables
	envPort         = "PORT"
	envDBHost       = "DB_HOST"
	envDBUser       = "DB_USER"
	envDBPass       = "DB_PASS"
	envDisableDB    = "DISABLE_DB"
	envRedisPort    = "REDIS_PORT"
	envRedisHost    = "REDIS_HOST"
	envRedisPass    = "REDIS_PASS"
	envDisableRedis = "DISABLE_REDIS"

	defaultPort      = "3000"
	defaultDBHost    = "127.0.0.1"
	defaultDBUser    = "dbuser"
	defaultDBPass    = "dbpass"
	defaultRedisPort = "6379"
	defaultRedisHost = "127.0.0.1"
	defaultRedisPass = ""
)

func openDB() (*sql.DB, error) {
	host := os.Getenv(envDBHost)
	if host == "" {
		host = defaultDBHost
	}

	user := os.Getenv(envDBUser)
	if user == "" {
		user = defaultDBUser
	}

	pass := os.Getenv(envDBPass)
	if pass == "" {
		pass = defaultDBPass
	}

	log.Printf("Opening DB connection to %s with user %s", host, user)
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/?tls=true", user, pass, host)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(context.TODO()); err != nil {
		return nil, err
	}

	return db, nil
}

var ctx = context.Background()

func redisClient() error {
	host := os.Getenv(envRedisHost)
	if host == "" {
		host = defaultRedisHost
	}

	port := os.Getenv(envRedisPort)
	if port == "" {
		port = defaultRedisPort
	}

	pass := os.Getenv(envRedisPass)
	if pass == "" {
		pass = defaultRedisPass
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         host + ":" + port,
		Password:     pass, // no password set
		DB:           0,    // use default DB
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			//Certificates: []tls.Certificate{cert}
		},
	})

	log.Printf("Opening Redis connection to %s:%s", host, port)

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(pong)

	// err := rdb.Set(ctx, "key", "value", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// val, err := rdb.Get(ctx, "key").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("key", val)

	// val2, err := rdb.Get(ctx, "key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }

	return err
}

func serveHTTP() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", hello)
	r.Get("/health", health)

	port := os.Getenv(envPort)
	if port == "" {
		port = defaultPort
	}
	log.Printf("Listening on port %s", port)
	return http.ListenAndServe(":"+port, r)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, Qmonus Value Stream!"))
}

func health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("healthy"))
}

func main() {
	if strings.ToLower(os.Getenv(envDisableDB)) != "true" {
		db, err := openDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	}

	if strings.ToLower(os.Getenv(envDisableRedis)) != "true" {
		if err := redisClient(); err != nil {
			log.Fatal(err)
		}
	}

	if err := serveHTTP(); err != nil {
		log.Fatal(err)
	}
}
