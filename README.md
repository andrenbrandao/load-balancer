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
❯ ./bin/be -p 8081
Listening for connections on 127.0.0.1:8081...

❯ ./bin/be -p 8081
Listening for connections on 127.0.0.1:8082...

❯ ./bin/be -p 8081
Listening for connections on 127.0.0.1:8083...
```

Now, execute the load balancer.

```bash
❯ ./bin/lb
Listening for connections on 127.0.0.1:8080...
```

You can simulate do a request by using cURL.

```bash
❯ curl http://localhost:8080
Hello From Backend Server
```

And to simulate multiple requests, you can use [wrk](https://github.com/wg/wrk).

```bash
❯ wrk http://localhost:8080 -t2 -c10 -d2
Running 2s test @ http://localhost:8080
  2 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   779.03us    1.41ms  16.49ms   94.93%
    Req/Sec     7.91k     1.68k   13.01k    90.24%
  32322 requests in 2.10s, 4.59MB read
Requests/sec:  15395.88
Transfer/sec:      2.19MB
```

## License

[MIT](LICENSE) © André Brandão
