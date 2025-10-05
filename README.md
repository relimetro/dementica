# dementica


# build and run
```
sudo docker-compose -f docker-compose.yaml build
sudo docker-compose -f docker-compose.yaml up
```

# test
`curl -X POST -d '{"name":"Bob"}' http://localhost:80/v1/say_hello`


