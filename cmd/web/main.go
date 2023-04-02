package main

import (
	"concurrent-subscriptions/data"
	"context"
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const webPort = "3000"

func main() {
	log.Printf("Starting application...")
	// load environment variables
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}

	// connect to the database
	db := initDB()
	defer db.Close()
	db.Ping()

	// create sessions
	session := initSession()

	// create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// create channels

	// create waitgroup
	wg := sync.WaitGroup{}

	// set up the application config
	app := Config{
		Session:  session,
		DB:       db,
		InfoLog:  infoLog,
		ErrorLog: errorLog,
		Wait:     &wg,
		Models:   data.New(db),
	}

	app.serve()

	// set up mail

	// listen for web connections

}

func (app *Config) serve() {
	srv := &http.Server{
		Addr:         ":" + webPort,
		Handler:      app.routes(),
		ErrorLog:     app.ErrorLog,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	app.InfoLog.Printf("Starting server on port %s", webPort)
	err := srv.ListenAndServe()
	if err != nil {
		app.ErrorLog.Fatal(err)
	}
}

// loadEnv loads environment variables from a .env file
func loadEnv() error {
	// load environment variables from .env file
	log.Printf("Loading environment variables from .env file...")
	if err := godotenv.Load(); err != nil {
		return err
	}

	return nil
}

// initDB initializes the database connection
func initDB() *sql.DB {
	// open the database connection
	log.Printf("Connecting to database...")
	conn := connectToDB()
	if conn == nil {
		log.Fatal("Could not connect to database")
	}

	return conn
}

// connectToDB attempts to connect to the database
func connectToDB() *sql.DB {
	attempts := 0

	dsn := os.Getenv("DB_DSN")

	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Printf("Could not connect to postgres: %s", err)
		} else {
			return conn
		}

		if attempts > 5 {
			return nil
		}

		log.Printf("Retrying postgres connection in 1 second...")
		time.Sleep(time.Second)
		attempts++
	}
}

// openDB opens a connection to the database
func openDB(dsn string) (*sql.DB, error) {
	log.Printf("Opening database connection: %s", dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// initSession initializes the session
func initSession() *scs.SessionManager {
	log.Printf("Initializing session...")
	gob.Register(data.User{})
	session := scs.New()

	redisPool := newRedisPool()
	session.Store = redisstore.New(redisPool)

	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

// newRedisPool initializes the Redis connection pool
func newRedisPool() *redis.Pool {
	log.Printf("Initializing Redis...")

	redisPool := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}

	// Ping the Redis server to check the connection
	conn := redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	return redisPool
}

func (app *Config) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	app.shutdown()
	os.Exit(0)
}

func (app *Config) shutdown() {
	// cleanup tasks
	app.InfoLog.Println("Cleaning up for shutdown...")

	// block until waitgroup is empty
	app.Wait.Wait()

	// close channels
	app.InfoLog.Println("Closing channels...")

	// close database connection

	// close mail connection

	// shutdown
	app.InfoLog.Println("Shutdown complete")
}
