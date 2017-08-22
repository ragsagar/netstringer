# Netstringer
Golang package to encode or decode stream of netstring

#### To install package:

`go get github.com/ragsagar/netstringer`

#### Examples:

##### To encode a data into netstring

```golang
//main.go
package main

import (
	"fmt"
	"github.com/ragsagar/netstringer"
)

func main() {
	input := []byte("hello world!")
	output := netstringer.Encode(input)
	fmt.Println(string(output)) // 12:hello world!,
}
```

```
$go run main.go
12:hello world!,
```

##### To decode a stream of incoming data into netstring

```golang
// main.go
package main

import (
	"fmt"
	"github.com/ragsagar/netstringer"
)

func main() {
        // To mock the incoming stream of data.
	inputs := []string{
		"12:hello world!,",
		"17:5:hello,6:world!,,",
		"12:hello ",
		"world!,",
	}
	decoder := netstringer.NewDecoder()
	go func() {
		for _, input := range inputs {
			decoder.FeedData([]byte(input))
		}
		close(decoder.DataOutput)
	}()

	for output := range decoder.DataOutput {
		fmt.Println(string(output))

	}
}
```
```
$go run main.go 
hello world!
5:hello,6:world!,
hello world!
```
