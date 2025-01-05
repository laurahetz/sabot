# Sabot - Efficient and Strongly Anonymous Bootstrapping

This is a prototype implementation of Sabot, a strongly anonymous bootstrapping protocol with bandwith-efficiency. 


> :warning: **Disclaimer**: This code is provided as an experimental implementation for testing purposes and should not be used in a productive environment. We cannot guarantee security and correctness.

## Code organization

The directories in this respository are:    

- **benchmarks**: Config files for our benchmarks
- **bootstrapping**: Anonymous bootstrapping protocol
  - includes server and client components
- **container**:
  - includes Containerfiles to simplify build and execution of this code base
- **lib/database**:
  - our key-value-store implementation for index- and keyword-based PIR based on [xorfilter](https://github.com/FastFilter/xorfilter)
- **lib/merkle**:
  - Merkle Tree implementation from [apir-code](https://github.com/dedis/apir-code), adapted for our protocol
- **lib/notify**:
  - XOR-Secret-Sharing implementation and construction of a notification matrix for our bootstrapping protocol
- **lib/pir**:
  - DPF-PIR implementation based on [checklist](https://github.com/dimakogan/checklist) and [dpf-go](https://github.com/dkales/dpf-go)
- **lib/utils**:
  - Merkle Tree implementation from [apir-code](https://github.com/dedis/apir-code), adapted for our protocol
- **modules**:
  - DPF implementation from [checklist](https://github.com/dimakogan/checklist)
- **proto**: 
  - Protobuf files for `bootstrapping`


## Requirements

- Golang 1.21.3
- [Protobuf](https://protobuf.dev/)
- [GRPC](https://grpc.io/)


## Containerized Build Process and Setup


This repo includes a containerized build and run environment.
The provided Makefile contains all provided commands.
While the following commands use Podman, a container orchestration software, compatible software, e.g. Docker, can be used instead. 

### 1. Build Container 

Just call the following command to build the container and generate certificates for authenticated channels between clients and server

```shell
make build
```

### 2. Generate Database Files 

To simplify benchmarking, our protocol components read in pre-generated files that contain the database and the protocol's public parameters for this database. 
To generate all database files required for our benchmarks at a predefined path (``./app/db/`) run

```shell
make db
```

To change the database parameters and the output path modify `./cmd/db-gen.sh`.

### 3. Run Benchmarks

The protocol benchmarks have 3 components: 2 servers and 1 benchmark driver. 
The benchmark driver allows running a series of benchmarks, initiated from the client side. 
It takes as input a `config.json` file that specidies the benchmarks to run.


1. Ensure `./app/benchmarks/configs.json` contains the desired benchmarks (For more details see `benchmarks/README.md`).
2. Change the `run` command in the `Makefile` to the desired paths and ports. 
3. Run
```shell
make run
```
to start the server components (as a background service), followed by the benchmark driver.
Please note (when not using the make command), that the server components need to be up and running before the benchmark driver can be started.

### 4. Cleanup

The results of the benchmarking can be found under the path specified in the config file. 
To remove the containers after the experiments, run 

```shell
make rm
```

### Helpful Container Commands

Display containers:
```shell
podman container ls
```

Check status of running container: 
```shell
podman logs -f <container name, e.g. s0, s1, bench>
```

Delete container:

```shell
podman container rm -f <container name, e.g. s0, s1, bench>
```


## Acknowledgements

This project uses several other projects as building blocks.

- The DPF-PIR code is based on [checklist](https://github.com/dimakogan/checklist) and [dpf-go](https://github.com/dkales/dpf-go)
- The binary fuse filter implementation is based on [xorfilter](https://github.com/FastFilter/xorfilter)
- The Merkle-Tree implementation is based on [apir-code](https://github.com/dedis/apir-code)