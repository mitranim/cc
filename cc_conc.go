package cc

import (
	"sync"
	"unsafe"
)

/*
Shortcut for `cc.Conc(funs).TryAll()`. Runs all given functions concurrently
and panics on error.
*/
func TryAll(funs ...func()) { Conc(funs).TryAll() }

/*
Shortcut for `cc.Conc(funs).All()`. Runs all given functions concurrently,
returning the resulting error.
*/
func All(funs ...func()) error { return Conc(funs).All() }

/*
Shortcut for `cc.Conc(funs).RunAll()`. Runs all given functions concurrently,
returning the resulting error slice.
*/
func RunAll(funs ...func()) Errs { return Conc(funs).RunAll() }

/*
Short for "concurrent". Features:

	* Run multiple functions concurrently.
	* Wait for all of them to finish.
	* Accumulate panics, returning multiple errors as `cc.Errs`.
	* Nearly no overhead for 0 or 1 functions.

Because it runs ALL functions to completion, special context support is not
required. Functions can closure some context externally.
*/
type Conc []func()

// Appends a function, but doesn't run it yet.
func (self *Conc) Add(fun func()) {
	if fun == nil {
		return
	}

	// Prevent a spurious escape that shows up in benchmarks.
	ptr := (*Conc)(noescape(unsafe.Pointer(self)))
	*ptr = append(*ptr, fun)
}

// Shortcut for running all functions with `Conc.All` and panicking on error.
func (self Conc) TryAll() {
	err := self.All()
	if err != nil {
		panic(err)
	}
}

/*
Runs the functions concurrently, waiting for all of them to finish. Catches
panics, converting them to `error`. If all functions succeed, returns nil,
otherwise returns an error. If there was only one function, its error is
returned as-is. If there were multiple functions, the underlying error value is
`cc.Errs`. The length and result positions in `cc.Errs` correspond to the
functions at the moment of calling this method.

For 0 or 1 functions, this has practically no overhead. It simply calls the only
function on the current goroutine and returns its result as-is.
*/
func (self Conc) All() error {
	switch self.size() {
	case 0:
		return nil

	case 1:
		return run(self.find())

	default:
		return self.runConc().Err()
	}
}

/*
Lower-level variant of `cc.Conc.All`. Instead of returning nil or non-nil
`error`, always returns `cc.Errs`, which may or may not be empty. The length
and result positions in `cc.Errs` correspond to the functions at the moment of
calling this method. Should be used ONLY if you want the error slice.
Otherwise, use `cc.Conc.All`.
*/
func (self Conc) RunAll() Errs {
	switch self.size() {
	case 0:
		return nil

	case 1:
		return Errs{run(self.find())}

	default:
		return self.runConc()
	}
}

func (self Conc) runConc() Errs {
	errs := make(Errs, len(self))
	var wg sync.WaitGroup

	for i, fun := range self {
		if fun == nil {
			continue
		}

		wg.Add(1)
		i, fun := i, fun
		go concRun(fun, &wg, &errs[i])
	}

	wg.Wait()
	return errs
}

func (self Conc) size() (count int) {
	for _, val := range self {
		if val != nil {
			count++
			if count >= 2 {
				return
			}
		}
	}
	return
}

func (self Conc) find() func() {
	for _, val := range self {
		if val != nil {
			return val
		}
	}
	return nil
}

func run(fun func()) (err error) {
	defer rec(&err)
	fun()
	return
}

func concRun(fun func(), wg *sync.WaitGroup, ptr *error) {
	defer wg.Add(-1)
	defer rec(ptr)
	fun()
}

func rec(ptr *error) { *ptr = toErr(recover()) }

// Borrowed from the standard	library.
func noescape(src unsafe.Pointer) unsafe.Pointer {
	out := uintptr(src)
	// nolint:staticcheck
	return unsafe.Pointer(out ^ 0)
}
