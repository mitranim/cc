package cc_test

import (
	"fmt"

	"github.com/mitranim/cc"
)

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
