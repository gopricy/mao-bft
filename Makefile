test:
	go test ./... -count=1

build:
	go build -o bin/demo demo/main.go

clean:
	rm -rf pst*
	rm -f *.json

init: build clean
	./bin/demo init 

leader: 
	clear
	./bin/demo -t=leader 1

follower:
	clear
	./bin/demo -t=follower $(filter-out $@,$(MAKECMDGOALS))
