## Overview

Short for "**C**on**C**urrent". Go tool for running multiple functions concurrently and collecting their results into an error slice. Tiny and dependency-free.

API docs: https://pkg.go.dev/github.com/mitranim/cc.

## Examples

```golang
import "github.com/mitranim/cc"

func ExampleAll() {
  err := cc.All(
    func() {
      fmt.Println(`running in background`)
    },
    func() {
      fmt.Println(`running in background`)
    },
  )
  fmt.Println(`done; no error:`, err == nil)

  // Output:
  // running in background
  // running in background
  // done; no error: true
}

func ExampleConc() {
  var conc cc.Conc

  conc.Add(func() {
    fmt.Println(`running in background`)
  })

  conc.Add(func() {
    fmt.Println(`running in background`)
  })

  fmt.Println(`starting`)

  err := conc.All()
  fmt.Println(`done; no error:`, err == nil)

  // Output:
  // starting
  // running in background
  // running in background
  // done; no error: true
}
```

## License

https://unlicense.org

## Misc

I'm receptive to suggestions. If this library _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
