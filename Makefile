test:
	go test ./... -count=1

build:
	go build -o bin/demo demo/main.go

clean:
	rm -rf pst*

init: build clean
	./bin/demo init 

leader: 
	./bin/demo -t=leader 1

follower:
	./bin/demo -t=follower $(filter-out $@,$(MAKECMDGOALS))
