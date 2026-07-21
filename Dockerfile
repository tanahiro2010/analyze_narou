FROM ubuntu:latest
LABEL authors="tanahiro2010"

RUN apt-get update && api-get install -y 

ENTRYPOINT ["top", "-b"]