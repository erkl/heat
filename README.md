Package **wire** is a composable, high-performance HTTP/1.1 toolkit.

The package defines a small number of distinct interfaces in an effort to
separate the different concerns in an HTTP stack. For example, implementations
of the `Transport` interface may execute round-trips over connections without
any knowledge of how those connections are established or managed; that should
be left entirely up to the `Dialer` underneath.


#### Benchmark

Let's be honest: constructing and parsing HTTP requests/responses isn't
realistically going to be the bottleneck in your programs. However, I feel that
it's important to show that breaking down an HTTP client into components isn't
going to result in worse performance. Because in fact, the exact opposite
happens.

Here are some benchmarks gathered by issuing a boatload of HTTP round-trips
over synthetic keep-alive connections, using go 1.4.2 on darwin/amd64.

```
net/http      100000      22805 ns/op      2166 B/op      36 allocs/op
wire          200000       8354 ns/op       860 B/op      26 allocs/op
```


#### License

```
Copyright (c) 2015, Erik Lundin.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE
OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.
```
