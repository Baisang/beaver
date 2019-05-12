clean:
	go clean

docker:
	go build
	docker build . -t baisang/beaver
	docker push baisang/beaver
