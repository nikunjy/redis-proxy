# Http Redis Proxy
A simple http proxy over redis [client](https://github.com/go-redis/redis)

This proxy lets you make three kinds of requests: 
* Get - This is a simple get as you would do on redis 
* Put - This is a simple put with no expiration on redis 
* CachedGet - Proxy maintains a local LRU cache and looks up stuff in there before looking into redis


## Run it locally
1. Build and run the containers
```
docker-compose build
docker-compose up
# You should be able to see the logs 
```

2. To put a value 
```
curl -X PUT  'localhost:8081/put?key=foo&val=bar'
# You should see:  'Wrote key foo and value bar'
```


3. You can now retrieve the value
```
curl 'localhost:8081/get?key=foo'
# You should see : 'bar'
```


4. If you did cached get it will not even go to redis
```
curl 'localhost:8081/cached_get?key=foo'
```

5. You can stop the service using `docker-compose kill`