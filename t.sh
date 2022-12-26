#!/usr/bin/env bash

# This is a simple demo scripts showing how to create, list and delete users.

printf "### Add users\n"
curl -d '{"name": "Jarmo", "age": 21}' -X POST "localhost:8080/user"
curl -d '{"name": "Anni", "age": 22}' -X POST "localhost:8080/user"
curl -d '{"name": "Timo", "age": 23}' -X POST "localhost:8080/user"
curl -d '{"name": "Sebastian", "age": 24}' -X POST "localhost:8080/user"
curl -d '{"name": "Daniel", "age": 25}' -X POST "localhost:8080/user"
curl -d '{"name": "Laura", "age": 26}' -X POST "localhost:8080/user"
curl -d '{"name": "Cecilia", "age": 27}' -X POST "localhost:8080/user"

printf "\n### Get users\n"
curl -X GET "localhost:8080/user/"

printf "\n### Delete users\n"
curl -d '{"name": "Jarmo", "id": 1}' -X DELETE "localhost:8080/user"
curl -d '{"name": "Daniel", "id": 5}' -X DELETE "localhost:8080/user"
curl -d '{"name": "Anni", "id": 2}' -X DELETE "localhost:8080/user"
