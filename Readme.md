# Golang microservices Transhipment project

## Requirements

- Golang 1.15
- Docker installed

## Instructions
To spin up application locally you just need to run
```
> make start
```
It will build all the services and spin up everything in docker containers

DB is migrated every time ports service is started as didn't want to mount any additional files to local reviewer machine.

To stop
```
> make stop
```

To run the tests
```
> make test
```
Didn't have time to write proper tests coverage but at least gave the idea how would I do it if I have time.

To regenerate proto files
```
> make proto
```
For it to work you need to have `TRANSSHIPMENT_ROOT` env var set to the absolute path to the `transhipment` folder

## Details
Important to note implementation lacks of validation, testing, and many more due to the short time frame.

Current implementation uses around ~32MB of memory to parse large files. Tested on 200MB json file.

There are options to tune performance with the buffer, but I think it should be out of scope for the task.

Every service is configurable and exposed on its own port.

For the ports service you can configure which internal service to load as a grpc handler.

You can change database default username/password/dbname values in config.yaml file.

Routes exposed:

-    GET    `/ports` list all the ports in the db
-    POST   `/ports` create new port
-    GET    `/ports/{idOrSlug}` fetch port by id or slug
-    PUT    `/ports/{idOrSlug}` update port by id or slug (partial update supported)
-    DELETE `/ports/{idOrSlug}` delete port by id or slug
-    POST   `/upload-ports` upload ports data file as form data in `file` key

Exposed ports:

- ports service: `58001`
- restgateway service: `58000`
- postgres: `54322`
