.PHONY: clean

cmd/pinglog/pinglog: cmd/pinglog/main.go
	cd cmd/pinglog; CGO_ENABLED=0 go build

clean:
	rm -f cmd/pinglog/pinglog
	find . -name *.json | xargs rm -f
