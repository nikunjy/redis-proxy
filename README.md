# Http Redis Proxy
A simple http proxy over redis [client](https://github.com/go-redis/redis)

This proxy lets you make two kinds of requests: 
* Get - This is a simple get as you would do on redis. Proxy keeps around a local LRU cache and if a key can be found in the LRU then it returns the response directly without looking into the redis store.
* Put - This is a simple put with no expiration on redis 


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


4. You can stop the service using `docker-compose kill`