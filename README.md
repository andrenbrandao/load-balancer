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
./bin/be -p 8081
./bin/be -p 8081
```

Now, execute the load balancer.

```bash
./bin/lb
```

You can simulate do a request by using cURL.

```bash
curl http://localhost:8080
```

And to simulate multiple requests, you can use [wrk](https://github.com/wg/wrk).

```bash
wrk http://localhost:8080 -t2 -c10 -d2
```

## License

[MIT](LICENSE) © André Brandão
