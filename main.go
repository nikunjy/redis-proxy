package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/nikunjy/redis-proxy/proxy"
	"github.com/nikunjy/redis-proxy/store"
)

type Options struct {
	CacheSize     int `long:"cache-size" description:"cache size"`
	ServerPort    int `long:"proxy-port" descriptiont:"port which proxy listens on"`
	ExpirySeconds int `long:"expiry-seconds" desciption:"expiry seconds for cache entries"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var proxyOptions []proxy.Option
	if opts.CacheSize > 0 {
		proxyOptions = append(proxyOptions, proxy.WithCacheSize(opts.CacheSize))
	}
	if opts.ServerPort > 0 {
		proxyOptions = append(proxyOptions, proxy.ListenOn(opts.ServerPort))
	}
	if opts.ExpirySeconds > 0 {
		proxyOptions = append(
			proxyOptions,
			proxy.WithCacheTTL(time.Second*time.Duration(opts.ExpirySeconds)),
		)
	}
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	redis, err := store.NewRedis(redisURL)
	if err != nil {
		log.Fatal("Error making redis client", err)
	}
	server, err := proxy.New(redis, proxyOptions...)
	if err != nil {
		log.Fatal(err)
	}
	srv := server.HttpServer()

	// Start Server
	go func() {
		log.Println("Starting Server ", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
