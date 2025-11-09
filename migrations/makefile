DB_URL=postgres://postgres:postgres@localhost:5432/meetingbot?sslmode=disable
MIGRATIONS_DIR=.

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" reset

migrate-redo:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" redo

migrate-version:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" version

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql
