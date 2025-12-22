# dementica

# website

available at http://dementica.danigoes.online.com


# build and run

```
sudo docker-compose -f docker-compose.yaml build
sudo docker-compose -f docker-compose.yaml up
```



## news feed

news_feed (for updating news every day) is ran separately,

run `cd ./services/firestore; ./server` and `cd ./services/news_feed; python main.py`

will run itself every 24 hours.






# test
`curl -X POST -d '{"name":"Bob"}' http://localhost:80/v1/say_hello`



# requirements

`./firebase.json` for user_service and firestore

`services/vertexai/*.json` serviceAccount for vertexAI

googleapis are required to build protofiles



# documentation

## services
consists of 6 services,
- firestore is used to interact with firestore db from firebase, as well as acting as external api
- frontend is a react app that serves web frontend
- keras is used to evaluate lifestyle questionnaires using local model, it is called by firestore using grpc
- vertexai similar to keras, but for the online api for vertexai used for transcription analysis
- news_feed is not part of docker-compose and is ran separately, used to generate news articles and adds them to database by calling firestore
- user_service provides functionality to login and register new accounts, and query for information


## database
database has 4 collections
- News: stores daily news for Patients and Doctors
- TestResults: stores tests uploaded by patients and the calculated risk score
	(date, data, RiskScore, Type(Lifestyle,Transcription,MMSE) UserID)
- Users: stores user information
	Email, Name, Type (Patient,Doctor,Admin)
	Patients also have:
		(DoctorID, HasDementia(Unknown,Positive,Negative), RiskScore)
-user_relations: stores relationships between users

## url
all http requests that start with "/v1/" are sent to backend api, all other requests are sent to webFrontend


