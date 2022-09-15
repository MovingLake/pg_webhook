build_latest:
	docker build -t ghcr.io/movinglake/pg_webhook:latest .

push_latest:
	docker push ghcr.io/movinglake/pg_webhook:latest

build_version:
	docker build -t ghcr.io/movinglake/pg_webhook:$(VERSION) .

push_version:
	docker build -t ghcr.io/movinglake/pg_webhook:$(VERSION) .