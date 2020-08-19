APP_EXECUTABLE?=./bin/customer
RELEASE?=0.1
MIGRATIONS_IMAGENAME?=arahna/customer-service-migrations:v$(RELEASE)
IMAGENAME?=arahna/customer-service:v$(RELEASE)

.PHONY: clean
clean:
	rm -f ${APP_EXECUTABLE}

.PHONY: build
build: clean
	docker build -t $(MIGRATIONS_IMAGENAME) -f DockerfileMigrations .
	docker build -t $(IMAGENAME) .

.PHONY: release
release:
	git tag v$(RELEASE)
	git push origin v$(RELEASE)