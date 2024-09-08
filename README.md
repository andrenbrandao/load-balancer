# Load Balancer - Coding Challenge #5

Load balancer implementation with Go. It is a solution to [John Crickett's Coding Challenges](https://codingchallenges.fyi/challenges/challenge-load-balancer).

![Terminal Animation](.github/terminal-animation.gif)

There are 2 main files:

- [lb](./cmd/lb/main.go) is the load balancer and it implements a Round Robin algorithm, sending requests to three different servers at ports 8081, 8082 and 8083.
- [be](./cmd/be/main.go) is a simple backend server written in Go to answer the GET requests.

## How to build

Execute `make build` to create the executables. It will be saved in the `bin/` directory.

## How to run

Execute three different web servers.

```bash
./bin/be -p 8081
./bin/be -p 8082
./bin/be -p 8083
```

Now, execute the load balancer.

```bash
./bin/lb
```

You can use cURL to make a request to our load balancer.

```bash
curl http://localhost:8080
```

And to simulate multiple requests, you can use [wrk](https://github.com/wg/wrk).

```bash
wrk http://localhost:8080 -t2 -c10 -d2
```

## How to run the tests

Execute `make test`

You should see a result matching something like:

```bash
--- PASS: TestHandlesMultipleClientsAtSameTime (1.00s)
PASS
ok      github.com/andrenbrandao/load-balancer/test     5.021s
```

## License

[MIT](LICENSE) © André Brandão
