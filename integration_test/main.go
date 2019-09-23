package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	buildCmd = `docker-compose build --build-arg EXPIRY=5`
	startCmd = `docker-compose up -d`
	killCmd  = `docker-compose kill`

	host = `localhost:8081`
	key  = `foo`
)

func serializeMap(m map[string]interface{}) string {
	data, err := json.Marshal(m)
	if err != nil {
		log.Println("Error marshalling the map", m)
		os.Exit(1)
	}
	return string(data)
}

func testPut(key, val string) error {
	uri := &url.URL{
		Scheme:   "http",
		Host:     host,
		Path:     "put",
		RawQuery: fmt.Sprintf("key=%s&val=%s", key, val),
	}
	log.Println("Making the put request", uri.String())
	req, err := http.NewRequest("PUT", uri.String(), nil)
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("http status should have been ok")
	}
	return nil
}

func testKey(key string, val string, found bool, fromCache bool) error {
	log.Println("Testing the key get from proxy, expecting state ", serializeMap(map[string]interface{}{
		"key":          key,
		"expected_val": val,
		"found":        found,
		"from_cache":   fromCache,
	}))
	uri := &url.URL{
		Scheme:   "http",
		Host:     host,
		Path:     "get",
		RawQuery: fmt.Sprintf("key=%s", key),
	}
	resp, err := http.Get(uri.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !found {
		if resp.StatusCode != http.StatusNotFound {
			return errors.New("Code should have been not found")
		}
		return nil
	}

	if string(data) != val {
		return errors.New("Response value does not match")
	}

	cachedHeader := resp.Header.Get("cached-value")
	if cachedHeader != strconv.FormatBool(fromCache) {
		return fmt.Errorf("Cached value header unexepcted, expected %s, actual %s", strconv.FormatBool(fromCache), cachedHeader)
	}
	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func randString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func runCommand(cmdStr string) {
	log.Println("Running the command ", cmdStr)
	splits := strings.Split(cmdStr, " ")
	cmd := exec.Command(splits[0], splits[1:]...)
	err := cmd.Run()
	if err != nil {
		log.Println("Error running the running command ", err)
		os.Exit(1)
	}
}

func bailOnErr(err error) {
	if err != nil {
		log.Println("Unexpected error", err)
		os.Exit(1)
	}
}

func main() {
	runCommand(buildCmd)
	runCommand(startCmd)
	defer runCommand(killCmd)

	rndKey := randString(10)
	fmt.Println("Testing out the key ", rndKey)
	bailOnErr(testKey(rndKey, "", false, false))
	bailOnErr(testPut(rndKey, "foobar"))
	// first get is cached
	bailOnErr(testKey(rndKey, "foobar", true, true))
	// second get is cached too
	bailOnErr(testKey(rndKey, "foobar", true, true))
	time.Sleep(time.Second * 11)
	bailOnErr(testKey(rndKey, "foobar", true, false))
	log.Println("Successful integration test")
}
