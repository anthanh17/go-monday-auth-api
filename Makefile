# ----------------------------------------- Setup create database ------------------------------------------------------
# Setup postgres database docker
createdbcontainer:
	docker run --name monday-auth-api -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=abc123 -d postgres:12-alpine

createdb:
	docker exec -it monday-auth-api createdb --username=root --owner=root monday_auth

dropdb:
	docker exec -it monday-auth-api dropdb monday_auth
# --------------------------------------------------------------------------------------------------------------------------
# -------------------------------------- Read file schema sql crete or update database --------------------------------------
# Migarte database all
migrateup:
	migrate -path db/migration -database "postgresql://root:abc123@localhost:5432/monday_auth?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:abc123@localhost:5432/monday_auth?sslmode=disable" -verbose down

# Migarte database lastest
migrateup1:
	migrate -path db/migration -database "postgresql://root:abc123@localhost:5432/monday_auth?sslmode=disable" -verbose up 1

migratedown1:
	migrate -path db/migration -database "postgresql://root:abc123@localhost:5432/monday_auth?sslmode=disable" -verbose down 1
# --------------------------------------------------------------------------------------------------------------------------
# ---------------------------------- Define schema dabase and define query sqlc generate code golang -----------------------
# create file config sqlc.yaml
sqlcinit:
	sqlc init

# sqlc gen code golang
sqlc:
	sqlc generate
# --------------------------------------------------------------------------------------------------------------------------
# Unit test
test:
	go test -v -cover ./...

# Start server http
server:
	go run main.go

# Mock
mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/anthanh17/simplebank/db/sqlc Store

.PHONY: createdbcontainer createdb dropdb migrateup migratedown sqlcinit sqlc test server mock migratedown1 migrateup1
