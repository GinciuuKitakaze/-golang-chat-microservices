up:
	docker compose up -d --build

down:
	docker compose down

restart:
	docker compose down
	docker compose up -d --build

user-up:
	docker compose up -d --build user postgres-user

user-down:
	docker compose stop user postgres-user

auth-up:
	docker compose up -d --build auth postgres-auth

auth-down:
	docker compose stop auth postgres-auth

postgres-up:
	docker compose up -d postgres-user postgres-auth

postgres-down:
	docker compose stop postgres-user postgres-auth

redis-up:
	docker compose up -d redis-auth

redis-down:
	docker compose stop redis-auth