package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/jessevdk/go-flags"
	"github.com/nikunjy/redis-proxy/proxy"
	"github.com/nikunjy/redis-proxy/store"
)

func ExampleNewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
}

func ExampleClient(client *redis.Client) {
	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
}

type Options struct {
	Redis         string `long:"redis" description:"redis address" default:"localhost:6379" required:"true"`
	CacheSize     int    `long:"cache-size" description:"cache size"`
	ServerPort    int    `long:"proxy-port" descriptiont:"port which proxy listens on"`
	ExpirySeconds int    `long:"expiry-seconds" desciption:"expiry seconds for cache entries"`
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
	redis, err := store.NewRedis(opts.Redis)
	if err != nil {
		log.Fatal("Error making redis client", err)
	}
	server, err := proxy.New(redis, proxyOptions...)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.ListenAndServe())
}
