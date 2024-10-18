run-server:
	go run cmd/server/main.go

run-testing-script:
	go run cmd/request_script/main.go --chats 100 --conns 15