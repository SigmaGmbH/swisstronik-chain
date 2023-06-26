# JSON RPC benchmark

## Install [`drill`](https://github.com/fcsonline/drill) with dependencies and run the benchmark

* Install [`cargo`](https://doc.rust-lang.org/cargo/getting-started/installation.html)

    ```bash
    $ curl https://sh.rustup.rs -sSf | sh
    $ source $HOME/.cargo/env
    ```

* Install dependencies

    ```bash
    $ sudo apt install -y build-essential pkg-config libssl-dev
    ```

* Clone this repo

    ```bash
    $ git clone https://github.com/near/jsonrpc-benchmark
    ```

* Navigate to the folder and run `drill`

    ```bash
    $ cd jsonrpc-benchmark/
    $ chmod a+x run.sh
    $ URL="http://localhost:3030" ./run.sh # to run each method sequentially
    $ drill --benchmark benchmark.yml --stats # to run the methods concurrently
    ```