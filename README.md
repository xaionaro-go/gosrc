[![GoDoc](https://godoc.org/github.com/xaionaro-go/gosrc?status.svg)](https://pkg.go.dev/github.com/xaionaro-go/gosrc?tab=doc)
[![go report](https://goreportcard.com/badge/github.com/xaionaro-go/gosrc)](https://goreportcard.com/report/github.com/xaionaro-go/gosrc)
<p xmlns:dct="http://purl.org/dc/terms/" xmlns:vcard="http://www.w3.org/2001/vcard-rdf/3.0#">
  <a rel="license"
     href="http://creativecommons.org/publicdomain/zero/1.0/">
    <img src="http://i.creativecommons.org/p/zero/1.0/88x31.png" style="border-style: none;" alt="CC0" />
  </a>
  <br />
  To the extent possible under law,
  <a rel="dct:publisher"
     href="https://github.com/xaionaro-go/rpn">
    <span property="dct:title">Dmitrii Okunev</span></a>
  has waived all copyright and related or neighboring rights to
  <span property="dct:title">gosrc</span>.
This work is published from:
<span property="vcard:Country" datatype="dct:ISO3166"
      content="IE" about="https://github.com/xaionaro-go/gosrc">
  Ireland</span>.
</p>


# About

This package just simplifies working with `go/*` packages to parse a source code.
Initially the package was written to simplify code generation.

# Quick start

```go
package main

import (
    "fmt"
    "go/build"

    "github.com/xaionaro-go/gosrc"
)

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dir, err := gosrc.OpenDirectoryByPkgPath(&build.Default, "github.com/xaionaro-go/gosrc", false, false, false, nil)
	assertNoError(err)

	if len(dir.Packages) != 1 {
		panic("expected one package")
	}
	pkg := dir.Packages[0]

	for _, file := range pkg.Files {
		for _, s := range file.Structs() {
			fields, err := s.Fields()
			assertNoError(err)
			fmt.Println("amount of fields:", len(fields), ";\t amount of methods:", len(s.Methods()), ";  \tstruct name:", s.Name())
		}
	}
}
```
The output is:
```
amount of fields: 2 ;    amount of methods: 0 ;         struct name: Directory
amount of fields: 2 ;    amount of methods: 1 ;         struct name: ErrPackageNotFound
amount of fields: 5 ;    amount of methods: 7 ;         struct name: Field
amount of fields: 2 ;    amount of methods: 0 ;         struct name: TypeNameValue
amount of fields: 3 ;    amount of methods: 12 ;        struct name: File
amount of fields: 1 ;    amount of methods: 0 ;         struct name: Func
amount of fields: 6 ;    amount of methods: 5 ;         struct name: Package
amount of fields: 3 ;    amount of methods: 5 ;         struct name: Struct
```