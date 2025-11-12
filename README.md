# dementica


# build and run
```
sudo docker-compose -f docker-compose.yaml build
sudo docker-compose -f docker-compose.yaml up
```


# test
`curl -X POST -d '{"name":"Bob"}' http://localhost:80/v1/say_hello`


# requirements

`./firebase.json` for user_service and firestore

`services/vertexai/*.json` serviceAccount for vertexAI

