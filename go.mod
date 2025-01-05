module sabot

go 1.21.3

replace github.com/dkales/dpf-go => ./modules/dpf-go

require (
	github.com/dkales/dpf-go v0.0.0-20210304170054-6eae87348848
	github.com/lukechampine/fastxor v0.0.0-20210322201628-b664bed5a5cc
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.67.0
	google.golang.org/protobuf v1.34.2
	lukechampine.com/blake3 v1.3.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
