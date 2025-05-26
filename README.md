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

This repo includes a containerized build and run environment.
The provided Makefile contains all required commands.
While the following commands use Podman, a container orchestration software, compatible software, e.g. Docker, can be used instead.

- Podman 
  - An equivalent container orchestration software, e.g., Docker, can be used, but for this, the commands in `Makefile` need to be adapted accordingly
- Make (optional as commands from the `Makefile` can be called directly)
- OpenSSL (tested with 3.5.0)

Building our prototype (inside the container) requires
- Golang 1.21.3
- [Protobuf](https://protobuf.dev/)
- [GRPC](https://grpc.io/)

## Containerized Build and Run Environment

We provide make commands to simplify the build and run commands for the benchmarking of our prototype.

Running the benchmarking will start start three containers: two server containers and one client container.

### 1. Build Container

To build the container and generate certificates for authenticated channels between clients and server run

```shell
make build
```

### 2. Generate Database Files 

To simplify benchmarking, our protocol components read in pre-generated files that contain the database and the protocol's public parameters for this database.
To generate all database files required for our benchmarks at a predefined path (`./app/db/`) run

```shell
make db
```

To change the database parameters and the output path modify `./cmd/db-gen.sh`.

### 3. Generate Benchmark Configurations

The benchmarking suite takes as input a `.json` file containing the descriptions of all benchmarks to run.

The script `cmd/config-gen.sh` helps with the creation of config files for different purposes.
It takes as interactive input the number of threads the simultated multi-client benchmarks should run on. Set this based on your available hardware.
It outputs the config files listed below
  - Full set of benchmarks: `app/benchmarks/configs_full.json` contains all benchmark configurations to run the experiments described in our paper. 
  - Small set of benchmarks: `app/benchmarks/configs_small.json` contains a reduced set of benchmark configurations to run experiments in shorter time and on hardware with less memory.
  - Test set of benchmarks: `app/benchmarks/configs_test.json` contains basic test configurations.

```shell
./cmd/config-gen.sh
```


### 4. Run Benchmarks

The protocol benchmarks have 3 components: 2 servers and 1 benchmark driver. 
The benchmarking suite allows running a series of benchmarks, initiated from the client side.
It takes as input a JSON file that specidies the benchmarks to run.

1. Ensure `./app/benchmarks/` contains the desired benchmarks (For more details see [the previous section](#generate-benchmark-configurations) and `benchmarks/README.md`).
2. Run the Make command for the according benchmark size (`make run-full`, `make run-small`, `make run-test`)
  ```shell
  # For the small benchmark set run
  make run-small
  ```
to start the server components (as a background service), followed by the benchmark driver.
Please note (when not using the make command), that the server components need to be up and running before the benchmark driver can be started.
3. **Wait for the benchmarks to finish**. The client container will stop running once it finished the calculations or if an error occured. Check the container logs for this `podman logs -f bench`. 

4. Once the client container stopped, remove all benchmarking containers using the command 
```shell
make rm
```

### 5. Cleanup

The results of the benchmarking can be found under the path specified in the config file (default: `./app/benchmark/results.csv`).
To remove the containers after the experiments, run 

```shell
make rm
```

### 6. Evaluate Results

The folder `eval` contains Python scripts to process the benchmarking results and to generate the tables displayed in our paper.
These scripts reqiure `numpy, pandas`. We recommend the use of a virtual environment to install these.
To do so run the following commands from the repository root:

```shell
python -m venv venv  
source venv/bin/activate
pip install numpy pandas
```

Run the evaluation scripts in this virtual environment.

The script `eval/eval.py` takes as input the benchmarking result file (csv) and processes it (incl. taking the mean value of the specified number of benchmark repetitions) to output the results of different experiments.
The results are stored in multiple files based on the specified output file prefix. 

```shell
python eval/eval.py <path to benchmark result csv file> <output file prefix> <number of repetitions>
```

The script `eval/eval-table.py` generates the tex table content for our paper based on files generated by the `eval.py` script and saved under the specified file prefix. This prefix is used here as the input. 

```shell
python eval/eval-table.py <input file prefix> <output file prefix>
```

Example for the small benchmarking suite:
```shell
python eval/eval.py ./app/benchmarks/results.csv result 5
python eval/eval-table.py result result
```

### Troubleshooting 

**Check Container Logs**

If anything does not work as expected, check the logs of the container(s) for errors with (see [Helpful Container Commands](helpful-container-commands)).
For the server container with name `s0` use the command `podman logs -f s0`. Use the names `s1, bench` for the other server and client container respectively.


**Container name in use**

After running benchmarks please note that the containers need to manually be shut off and removed.
Use the command `make rm` to remove the containers created for benchmarking.
All benchmark executions use the same container names, so no new containers can be started if previous ones have not been removed and the following error message will show: 
```shell
Error: creating container storage: the container name "s0" is already in use by <some id>. You have to remove that container to be able to reuse that name: that name is already in use, or use --replace to instruct Podman to do so.
```

**File Write Permissions**

When testing on Fedora Server, we ran into issues with file permissions during DB generation:
```shell
error reading db file:open /app/db/db_10_32_32_false.ipir: permission denied
```
These issues can be resolved by disabling SELinux using the command `sudo setenforce 0`. However, please be aware of the security implications of this change.




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