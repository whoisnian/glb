// Package logger implements a simple logger based on log/slog.
//
// NanoHandler formats slog.Record as a sequence of value strings without attribute keys to minimize log length.
//
//	2023/08/16 00:35:15 [I] Service httpd started: <http://127.0.0.1:9000>
//	2023/08/16 00:35:15 [D] Use default cache dir: /root/.cache/glb
//	2023/08/16 00:35:15 [W] Failed to check for updates: i/o timeout
//	2023/08/16 00:35:15 [E] dial tcp 127.0.0.1:9000: connect: connection refused
//	2023/08/16 00:35:15 [F] mkdir /root/.config/glb: permission denied
//
//	2023/08/16 00:35:15 [I] REQUEST 10.0.3.201 GET /status R5U3KA5C-42
//	2023/08/16 00:35:15 [I] FETCH Fetch upstream metrics http://127.0.0.1:8080/metrics 3 R5U3KA5C-42
//	2023/08/16 00:35:15 [I] RESPONSE 200 4 10.0.3.201 GET /status R5U3KA5C-42
//
// TextHandler formats slog.Record as a sequence of key=value pairs separated by spaces and followed by a newline.
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Service httpd started: <http://127.0.0.1:9000>"
//	time=2023-08-16T00:35:15+08:00 level=DEBUG msg="Use default cache dir: /root/.cache/glb"
//	time=2023-08-16T00:35:15+08:00 level=WARN msg="Failed to check for updates: i/o timeout"
//	time=2023-08-16T00:35:15+08:00 level=ERROR msg="dial tcp 127.0.0.1:9000: connect: connection refused"
//	time=2023-08-16T00:35:15+08:00 level=FATAL msg="mkdir /root/.config/glb: permission denied"
//
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" tag=REQUEST ip=10.0.3.201 method=GET path=/status tid=R5U3KA5C-42
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="Fetch upstream metrics" tag=FETCH url=http://127.0.0.1:8080/metrics duration=3 tid=R5U3KA5C-42
//	time=2023-08-16T00:35:15+08:00 level=INFO msg="" tag=RESPONSE code=200 duration=4 ip=10.0.3.201 method=GET path=/status tid=R5U3KA5C-42
//
// JsonHandler formats slog.Record as line-delimited JSON objects.
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Service httpd started: <http://127.0.0.1:9000>"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"DEBUG","msg":"Use default cache dir: /root/.cache/glb"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"WARN","msg":"Failed to check for updates: i/o timeout"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"ERROR","msg":"dial tcp 127.0.0.1:9000: connect: connection refused"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"FATAL","msg":"mkdir /root/.config/glb: permission denied"}
//
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","tag":"REQUEST","ip":"10.0.3.201","method":"GET","path":"/status","tid":"R5U3KA5C-42"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"Fetch upstream metrics","tag":"FETCH","url":"http://127.0.0.1:8080/metrics","duration":3,"tid":"R5U3KA5C-42"}
//	{"time":"2023-08-16T00:35:15.208873091+08:00","level":"INFO","msg":"","tag":"RESPONSE","code":200,"duration":4,"ip":"10.0.3.201","method":"GET","path":"/status","tid":"R5U3KA5C-42"}
package logger
