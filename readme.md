# Tiny-Redis

![](badge.svg) [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Yet another Godis, inspired by [Godis](https://github.com/archeryue/godis) && [Build Your Own Redis](https://build-your-own.org/redis/).


## References

### The SO_REUSEPORT socket option

https://lwn.net/Articles/542629/

### RESP Protocol

https://redis.io/docs/reference/protocol-spec/

#### Response

Using CRLF("\r\n") as separator.

- `+`: simple string, example: "+OK\r\n"

- `-`: simple error, example: "-Error message\r\n"

- `:`: integers, example: ":100\r\n"

- `$`: bulk string, example: "$6\r\nfoobar\r\n"

- `*`: array, examle: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"


#### Inline Command

"set key val\r\n"

Don't support key or value contains whitespace!

#### MultiBulk Command

"*3\r\n$3\r\nset\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

