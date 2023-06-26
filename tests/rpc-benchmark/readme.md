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

* Navigate to the folder and run `drill`

    ```bash
    $ URL="http://localhost:3030" drill --quiet --benchmark benchmark.yml --stats
    ```