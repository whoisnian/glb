# bench

## 20220720
```
#GithubAPI Routes: 203
   HttpRouter: 37136 Bytes
   Httpd: 65936 Bytes

=== RUN   TestRouters
--- PASS: TestRouters (0.00s)
goos: linux
goarch: amd64
pkg: github.com/whoisnian/glb/bench
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkHttpRouter_GithubAll
BenchmarkHttpRouter_GithubAll      62625             16653 ns/op               0 B/op          0 allocs/op
BenchmarkHttpd_GithubAll
BenchmarkHttpd_GithubAll           12733             94213 ns/op           89536 B/op       1113 allocs/op
PASS
ok      github.com/whoisnian/glb/bench  3.412s
```

## 20220721
```
#GithubAPI Routes: 203
   HttpRouter: 37136 Bytes
   Httpd: 65728 Bytes

=== RUN   TestRouters
--- PASS: TestRouters (0.00s)
goos: linux
goarch: amd64
pkg: github.com/whoisnian/glb/bench
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkHttpRouter_GithubAll
BenchmarkHttpRouter_GithubAll      61336             16780 ns/op               0 B/op          0 allocs/op
BenchmarkHttpd_GithubAll
BenchmarkHttpd_GithubAll           15352             78358 ns/op           74080 B/op        910 allocs/op
PASS
ok      github.com/whoisnian/glb/bench  3.243s
```

## 20220723
```
#GithubAPI Routes: 203
   HttpRouter: 37136 Bytes
   Httpd: 71344 Bytes

=== RUN   TestRouters
--- PASS: TestRouters (0.00s)
goos: linux
goarch: amd64
pkg: github.com/whoisnian/glb/bench
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkHttpRouter_GithubAll
BenchmarkHttpRouter_GithubAll      67580             17451 ns/op               0 B/op          0 allocs/op
BenchmarkHttpd_GithubAll
BenchmarkHttpd_GithubAll           45955             25345 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/whoisnian/glb/bench  2.799s
```

## 20220724
```
#GithubAPI Routes: 203
   HttpRouter: 37136 Bytes
   Gin: 58792 Bytes
   Httpd: 71344 Bytes

goos: linux
goarch: amd64
pkg: github.com/whoisnian/glb/bench
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkHttpRouter_GithubAll      70322             16652 ns/op               0 B/op          0 allocs/op
BenchmarkGinRouter_GithubAll       56919             21207 ns/op               0 B/op          0 allocs/op
BenchmarkHttpd_GithubAll           45211             27776 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/whoisnian/glb/bench  4.353s
```

## 20230821
```
#GithubAPI Routes: 203
   HttpRouter: 37136 Bytes
   Gin: 58792 Bytes
   Httpd: 84400 Bytes

goos: linux
goarch: amd64
pkg: github.com/whoisnian/glb/bench
cpu: AMD Ryzen 7 3700X 8-Core Processor             
BenchmarkHttpRouter_GithubAll      75914             15665 ns/op               0 B/op          0 allocs/op
BenchmarkGinRouter_GithubAll       59564             20646 ns/op               0 B/op          0 allocs/op
BenchmarkHttpd_GithubAll           44200             27045 ns/op               0 B/op          0 allocs/op
PASS
ok      github.com/whoisnian/glb/bench  4.281s
```
