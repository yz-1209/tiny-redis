## The SO_REUSEPORT socket option

https://lwn.net/Articles/542629/

## RESP Protocol

https://redis.io/docs/reference/protocol-spec/

CRLF "\r\n".

- `+`: simple string, example: "+OK\r\n"

- `-`: simple error, example: "-Error message\r\n"

- `:`: integers, example: ":100\r\n"

- `$`: bulk string, example: "$6\r\nfoobar\r\n"

- `*`: array, examle: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"


### Inline

"set key val\r\n" 

cons: don't support key or value contains whitespace!

### MultiBulk

"*3\r\n$3\r\nset\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

