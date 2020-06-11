# some useful variables
VERBOSE = 
GO_ENV = 
GO_DOCKER_ENV = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
IMAGE_NAME = avoxi-test
IMAGE_REGISTRY = bigfatmiddle
#registry-1.docker.io

MAJOR_GOVERSION=$(shell go version | grep -o '1\.[0-99]*' )
PACKAGES = $(shell find ./ -type d -not -path '*/\.*')


# create binary
.PHONY: quick-local-binary
quick-local-binary:
	go build -o main
.PHONY: binary
binary:
	$(GO_DOCKER_ENV) go build $(VERBOSE) -o main


# download dependencies for binary
.PHONY: deps
deps:
	export GO111MODULE=on
	$(GO_ENV) go get -d $(VERBOSE) ./...


# run unit tests
# get dependencies for gocov-xml and go-junit-report so we can send test results to devops-insights
.PHONY: test
test:
	$(GO_ENV) go get github.com/axw/gocov/gocov
	$(GO_ENV) go get github.com/AlekSi/gocov-xml
	$(GO_ENV) go get -u github.com/jstemmer/go-junit-report

	$(GO_ENV) go get -d -t $(VERBOSE) ./...
	$(GO_ENV) go vet ./...
	
	echo "GO version: $(MAJOR_GOVERSION)"

# go test -coverprofile option is supported when running multiple tests only starting at go version 1.10.

ifeq ($(shell expr $(MAJOR_GOVERSION) \< 1.10), 1)
	echo "mode: count" > coverage-all.out
	echo "" > unittest.out
	$(foreach pkg,$(PACKAGES),\
		$(GO_ENV) go test $(VERBOSE) -coverprofile=coverage.out -covermode=count $(pkg) >> unittest.out || echo "Error testing $(pkg)";\
		tail -n +2 coverage.out >> coverage-all.out || echo "No coverage for $(pkg)";)
else
	-$(GO_ENV) go test `go list ./... | grep -v testutils` $(VERBOSE) -coverpkg ./... -coverprofile coverage-all.out  > unittest.out
endif

	-cat unittest.out  #for logging to stdout
	-cat unittest.out | go-junit-report > unittest.xml
	-$(GO_ENV) gocov convert coverage-all.out | gocov-xml > coverage.xml

	cat unittest.out | grep "^FAIL" && exit 1 || true

# run a code scan
# todo: fix error handling and remove the exclude flag below
.PHONY: scan
scan:
	$(GO_ENV) go get github.com/securego/gosec/cmd/gosec
	CGO_ENABLED=0 gosec -exclude=G104 ./...


# build the docker image
.PHONY: image
image: binary
	docker build -t $(IMAGE_NAME):latest .

# tag and push official image
.PHONY: image-push
image-push:
	git diff-index --quiet HEAD -- #error if uncommitted changes
	$(eval IMAGE_TAG = $(shell git rev-parse HEAD))
	docker tag $(IMAGE_NAME):latest $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


# push image and deploy to OSSDev personal namespace
# todo: run kdep with custom tag
.PHONY: deploy
deploy:
	$(eval IMAGE_TAG = $(shell whoami|cut -d@ -f1)-$(shell date +%s))
	docker tag $(IMAGE_NAME):latest $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(IMAGE_REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)


# official cicd step ran from Jenkins
.PHONY: cicd-full
cicd-full: cicd-full-setup deps test scan image clean
.PHONY: cicd-full-setup
cicd-full-setup:
	$(eval VERBOSE = -v)
	$(eval GO_ENV = $(GO_DOCKER_ENV))
	$(eval GO111MODULE=on)
	./cicd-setup.sh || true

.PHONY: cicd-test
cicd-test: cicd-full-setup deps scan test clean


# build the binary and publishes it to the developer's namespace in OSSDev, in the future also calls kdep to deploy
.PHONY: cicd
cicd: image deploy


# remove artifacts and cleanup
.PHONY: clean
clean:
	rm main || true
	./cicd-cleanup.sh || true
