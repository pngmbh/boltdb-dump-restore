all: dump restore

dump restore: dump.go
	go build dump
	cp dump restore
