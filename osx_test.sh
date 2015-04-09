#!/bin/bash

boot2docker start

docker pull dockerfile/rethinkdb
docker pull dockerfile/rabbitmq


rabbitmq_instance=$(docker run -d -p 5672:5672 -p 15672:15672 dockerfile/rabbitmq)
rethink_instance=$(docker run -d -p 8080:8080 -p 28015:28015 -p 29015:29015 dockerfile/rethinkdb)
if [ $? -ne 0 ]; then
	echo "Failed to start dockerfile/rethinkdb"
	exit 1
fi


docker_ip=$(boot2docker ip)

RABBITMQ_HOST=${docker_ip} RETHINKDB_HOST=${docker_ip} go test ./...

docker kill ${rethink_instance}
docker kill ${rabbitmq_instance}
