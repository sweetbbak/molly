default:
	go build -o molly -ldflags='-s -w' ./cmd/molly

molly2:
	go build -o molly++ -ldflags='-s -w' ./cmd/molly2
