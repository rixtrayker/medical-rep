.PHONY: docker-up docker-down docker-logs docker-restart

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	$(MAKE) docker-down
	$(MAKE) docker-up
