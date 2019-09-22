package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	lru "github.com/hashicorp/golang-lru"

	"github.com/nikunjy/redis-proxy/store"
)

type proxyHandler struct {
	config Config
	cache  *lru.Cache
	store  store.Store
}

type cachedValue struct {
	val      string
	storedAt time.Time
}

func New(store store.Store, options ...Option) (*proxyHandler, error) {
	c := defaultConfig()
	for _, opt := range options {
		opt(c)
	}
	cache, err := lru.New(c.cacheSize)
	if err != nil {
		return nil, err
	}
	return &proxyHandler{
		config: *c,
		cache:  cache,
		store:  store,
	}, nil
}

func (p *proxyHandler) get(key string) (string, error) {
	val, err := p.store.Get(key)
	if err != nil {
		return "", err
	}
	p.cache.Add(key, cachedValue{val, time.Now()})
	return val, nil
}

func (p *proxyHandler) cachedGet(key string) (string, error) {
	val, ok := p.cache.Get(key)
	if !ok {
		return p.get(key)
	}
	cv, ok := val.(cachedValue)
	if !ok {
		// Lets delete from cache
		p.cache.Remove(key)
		return p.get(key)
	}
	if time.Now().Sub(cv.storedAt) > p.config.cacheTTL {
		return p.get(key)
	}
	return cv.val, nil
}

func processGetWithFn(
	w http.ResponseWriter,
	r *http.Request,
	getFn func(string) (string, error),
) {
	vals := r.URL.Query()
	key := vals.Get("key")
	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no key specified"))
		return
	}
	val, err := getFn(string(key))
	if err != nil {
		if err == store.ErrKeyNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("key %s not found", key)))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error getting key %s", key)))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))

}

func (p *proxyHandler) CachedGet(w http.ResponseWriter, r *http.Request) {
	processGetWithFn(w, r, p.cachedGet)
}

func (p *proxyHandler) Get(w http.ResponseWriter, r *http.Request) {
	processGetWithFn(w, r, p.get)
}

func (p *proxyHandler) Put(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	key := vals.Get("key")
	if len(key) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no key specified"))
		return
	}
	val := vals.Get("val")
	if len(val) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no val specified"))
		return
	}
	if err := p.put(key, val); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error writing key %s and value %s, error: %v", key, val, err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Wrote key %s and value %s", key, val)))
}

func (p *proxyHandler) put(key, val string) error {
	if err := p.store.Set(key, val); err != nil {
		return err
	}
	p.cache.Add(key, cachedValue{val, time.Now()})
	return nil
}

func (p *proxyHandler) ListenAndServe() error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/get", p.Get).Methods("GET")
	router.HandleFunc("/cached_get", p.CachedGet).Methods("GET")
	router.HandleFunc("/put", p.Put).Methods("PUT")
	return http.ListenAndServe(fmt.Sprintf(":%d", p.config.listenPort), router)
}
