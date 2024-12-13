package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/data"
	"github.com/LidoHon/LetsGOFurther-Greenlight.git/internal/jsonlog"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


const version = "1.0.0"

type config struct{
	port int
	env string
	db struct{
		dsn 			string
		maxOpenConns 	int
		maxIdleConns 	int
		maxIdleTime 	string
	}
	limiter struct{
		rps  float64
		burst int 
		enabled bool
	}
}


type application struct{
	config config
	logger *jsonlog.Logger
	models data.Models
}


func main(){
	err := godotenv.Load("../../.env")
	if err != nil{
		log.Fatalf("error loading the enviromental variable")
	}
	log.Println("Successfully loaded .env file")

	dsn := os.Getenv("DB_DSN")
	log.Printf("DB_DSN: %s", dsn)
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	

	flag.StringVar(&cfg.db.dsn, "db-dsn",dsn, "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

// rate limiter flages
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	if cfg.db.dsn == "" {
		log.Fatal("DB_DSN not set")
	}

	logger :=jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connected", nil)

	app :=&application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}


	srv:= http.Server{
		Addr: 			fmt.Sprintf(":%d", cfg.port),
		Handler: 		app.routes(),
		ErrorLog: 		log.New(log.Writer(), "", 0),
		IdleTimeout: 	time.Minute,
		ReadTimeout: 	10*time.Second,
		WriteTimeout:	10*time.Second,
	}

	logger.PrintInfo("starting server",map[string]string{
		"addr":srv.Addr,
		"env": cfg.env,
		
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)

}

func openDB(cfg config)(*sql.DB, error){
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil{
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)
	

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil{
		return nil, err
	}
	return db, nil
}