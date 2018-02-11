**heat** is a low-level HTTP/1.X library for Go. It defines minimal types
representing requests and responses, and provides functions for reading
and writing both headers and message bodies.

Being dumb by design, the package's functions don't try to massage invalid
requests/responses into sane ones; `WriteResponseHeader` will happily spit
out a `HTTP/33.-2` version string if that's what you ask it to do. As a
result these functions are highly efficient, completely predictable, and
probably a bit dangerous.

See [godoc.org](http://godoc.org/github.com/erkl/heat) for the specifics.


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
