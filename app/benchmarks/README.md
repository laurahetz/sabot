# Benchmarking Configurations

This folder contains the benchmarking configurations used to gather the results for our work.
The results of each benchmark are stored as a row in a csv file specified when running the benchmarks.

The benchmarks in the config files include

- DBsizeExp: 2^10, 2^12, 2^14, 2^16, 2^18
- Rates: 1, 5, 10
- keysize and valuesize: 32 byte
- Repetitions: 50

Bandwidth is in byte and runtime is in microseconds.


Each benchmark has the following parameters
```
{
  "Idx": 0,   # index of the executed client, can be used when manually running multiple clients
  "Dbfile": "/app/db/db_12_32_32_false", # path to DB file(s) in the container, please ensure file(s) exists. without ".ipir" or ".kwpir" ending.
  "RateR": 1, # receiver retrieval rate
  "RateS": 1, # sender retrieval rate
  "MultiClient": false, # run multiclient simulation
  "NumThreads": 1,  # number of threads to use in multiclient simulation
  "ResetServer": true,  # when running multiple bemchmarks on same DB file, the server can be reused and does not need a reset, the first benchmark needs be set to `true`
  "Repetitions": 50  # number of repetitions for this benchmark
}
```