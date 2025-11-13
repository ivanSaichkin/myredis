# myredis

myredis is a Go project that implements a simple Redis server with support for basic commands such as SET, GET, and DEL. It uses the Go standard library's TCP networking package to handle client connections and the `encoding/gob` package to serialize and deserialize data.

## Features

* Implements basic Redis commands such as SET, GET, DEL, HSET, HGET, HDEL, LLPUSH, LPOP, RPUSH, RPOP, SADD, SREM, SMEMBERS, and SMISMEMBER.
* Supports multiple databases with separate data storage for each database.
* Uses a hash table to store SET and GET values.
* Implements a list data structure using a linked list.
* Implements a set data structure using a bitset.
* Supports persistence using a binary file.
* Provides an option to enable AOF (append-only file) mode for persistence.
* Implements a configuration file with options for setting the server's address, port, and database.

## Installation Instructions

To use this project, you need to have Go installed on your system. Once you have Go installed, you can install the dependencies by running `go get` in the terminal:
```
$ go get -u ./...
```
This will download and install all the required dependencies for the project.

## Usage Examples

To start the Redis server, run the following command in the terminal:
```
$ go run cmd/server/main.go
```
This will start the Redis server on port 6379 by default. You can change the port number by providing a different value for the `port` flag when starting the server. For example, to start the server on port 8080, run the following command:
```
$ go run cmd/server/main.go --port=8080
```
Once the server is started, you can use a Redis client library or a tool like redis-cli to interact with the server. For example, to set a key-value pair in the Redis database, you can use the following command:
```
SET mykey "my value"
```
To get the value associated with the key `mykey`, you can use the following command:
```
GET mykey
```
To delete a key from the Redis database, you can use the following command:
```
DEL mykey
```
## Configuration

The configuration file for this project is located at `./internal/config/config.go`. You can modify the options in this file to change the behavior of the server. For example, you can change the port number or the database that the server uses by modifying the corresponding values in the `Config` struct.

## Project Structure

The project structure for this project is as follows:
```
.
├── cmd/server/main.go
│   └── Handles incoming client connections and dispatches commands to the appropriate handler functions.
├── data/dump.bin
│   └── Stores the Redis database in binary format.
├── go.mod
│   └── Declares the Go module dependencies for this project.
├── internal/command/executor.go
│   └── Handles command execution and delegates to the appropriate handler function.
├── internal/command/hashCommands.go
│   └── Implements the SET, GET, HSET, HGET, HDEL, and DEL commands for the Redis hash data type.
├── internal/command/listCommands.go
│   └── Implements the LLPUSH, LPOP, RPUSH, RPOP, and SADD commands for the Redis list data type.
├── internal/command/parser.go
│   └── Parses incoming client requests and extracts the relevant information such as command name and arguments.
├── internal/command/setCommands.go
│   └── Implements the SET, GET, DEL, HSET, HGET, HDEL, SADD, SREM, and SMISMEMBER commands for the Redis set data type.
├── internal/command/validator.go
│   └── Validates incoming client requests to ensure that they are well-formed and do not contain any syntax errors.
├── internal/config/config.go
│   └── Defines the Config struct and provides functions for reading and writing configuration options to a file.
├── internal/protocol/resp.go
│   └── Provides functions for encoding and decoding Redis protocol messages in RESP (REdis Serialization Protocol) format.
├── internal/protocol/resp_test.go
│   └── Contains test cases for the RESP protocol functions.
├── internal/server/handler.go
│   └── Handles incoming client connections and dispatches commands to the appropriate handler function.
├── internal/server/tcp.go
│   └── Implements the TCP server that listens for incoming client connections.
├── internal/storage/complex_test.go
│   └── Contains test cases for the Complex data structure.
├── internal/storage/hash.go
│   └── Implements the Hash data structure with a map of key-value pairs.
├── internal/storage/list.go
│   └── Implements the List data structure with a linked list.
├── internal/storage/memory.go
│   └── Provides functions for allocating and deallocating memory for Redis data structures.
├── internal/storage/persistence.go
│   └── Implements the Persistence functionality that saves the Redis database to disk periodically.
├── internal/storage/persistenceEngine.go
│   └── Provides functions for managing the persistence process, including saving and loading the Redis database from disk.
├── internal/storage/persistence_test.go
│   └── Contains test cases for the Persistence functionality.
├── internal/storage/set.go
│   └── Implements the Set data structure with a bitset.
├── internal/storage/storageEngine.go
│   └── Provides functions for managing the Redis database, including inserting and retrieving key-value pairs.
├── internal/storage/types.go
│   └── Defines the types used in the Redis data structures, such as String, Hash, List, Set, and Persistence.
└── README.md
    └── This file.
```
## Contributing Guidelines

We welcome contributions from the community. Before sending a pull request, please ensure that you have tested your changes using the testing framework provided with this project. Also, ensure that you have read and understood the contributing guidelines in the README file of this project.

## License

This project is licensed under the MIT License.