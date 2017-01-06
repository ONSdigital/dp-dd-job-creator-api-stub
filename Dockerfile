FROM ubuntu:16.04

WORKDIR /app/

COPY ./build/dp-dd-job-creator-api-stub .

ENTRYPOINT ./dp-dd-job-creator-api-stub
