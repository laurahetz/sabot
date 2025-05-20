build:
	./cmd/cert-gen.sh && podman build -f container/Containerfile . -t sabot

run:
	podman run -d -v ./app/db:/app/db --name s0 --network host  sabot /app/server -port=50051 && podman run -d -v ./app/db:/app/db --name s1 --network host sabot /app/server -port=50052 && podman run -d -v ./app/benchmarks:/app/benchmarks --name bench --network host sabot /app/benchmark -path /app/benchmarks/configs.json

rm:
	podman rm -f s0 s1 bench

db:
	./cmd/db-gen.sh