# MyRedis - Key-Value storage on Go

## Start

```bash
# start server
go run cmd/server/main.go

# connection with redis-cli
redis-cli -p 6379
```

## Usage Examples

### Hash operations

```bash
127.0.0.1:6379> HSET user:1 name Alice age 30 city "New York"
(integer) 3
127.0.0.1:6379> HGET user:1 name
"Alice"
127.0.0.1:6379> HGETALL user:1
1) "name"
2) "Alice"
3) "age"
4) "30"
5) "city"
6) "New York"
127.0.0.1:6379> HLEN user:1
(integer)
```

### List operations

```bash
127.0.0.1:6379> LPUSH mylist world
(integer) 1
127.0.0.1:6379> LPUSH mylist hello
(integer) 2
127.0.0.1:6379> LPOP mylist
"hello"
127.0.0.1:6379> RPUSH mylist "!"
(integer) 2
127.0.0.1:6379> LRANGE mylist 0 -1
1) "world"
2) "!"
```

### Set operations

```bash
127.0.0.1:6379> SADD tags go python java
(integer) 3
127.0.0.1:6379> SISMEMBER tags go
(integer) 1
127.0.0.1:6379> SMEMBERS tags
1) "go"
2) "python"
3) "java"
127.0.0.1:6379> SADD tags2 go rust
(integer) 2
127.0.0.1:6379> SINTER tags tags2
1) "go"
```

### type definition

```bash
127.0.0.1:6379> SET mystring "hello"
OK
127.0.0.1:6379> TYPE mystring
string
127.0.0.1:6379> HSET myhash field value
(integer) 1
127.0.0.1:6379> TYPE myhash
hash
```
