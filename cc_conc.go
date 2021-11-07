package cc

import "sync"

// Same as `make(cc.Conc, 0, count)` but can be syntactically cleaner.
func Make(count int) Conc { return make(Conc, 0, count) }

/*
Shortcut for `cc.Conc(...funs).TryAll()`. Runs all given functions concurrently
and panics on error.
*/
func TryAll(funs ...func()) { Conc(funs).TryAll() }

/*
Shortcut for `cc.Conc(...funs).All()`. Runs all given functions concurrently,
returning the resulting error.
*/
func All(funs ...func()) error { return Conc(funs).All() }

/*
Short for "concurrent". Features:

	* Run multiple functions concurrently.
	* Wait for all of them to finish.
	* Accumulate panics, returning multiple errors as `cc.Errs`.

Because it runs ALL functions to completion, special context support is not
required. Functions can closure some context externally.
*/
type Conc []func()

// Appends a function, but doesn't run it yet.
func (self *Conc) Add(fun func()) {
	if fun != nil {
		*self = append(*self, fun)
	}
}

// Shortcut for running all functions with
func (self Conc) TryAll() {
	err := self.All()
	if err != nil {
		panic(err)
	}
}

/*
Shortcut for `self.RunAll().Err()` that correctly converts the error slice to
the `error` interface, returning nil on complete success.

Runs the functions concurrently. If all succeed, returns nil. If there is at
least one failure, returns a non-nil error whose concrete value is `cc.Errs`.
The length and result positions in `cc.Errs` correspond to the functions at the
moment of calling this method.

Catches panics, converting them to `error` and storing them in the resulting
error slice. However, non-nil panics which don't implement `error` are not
supported. If such a panic is caught, this will re-panic in a background
goroutine and crash your app.
*/
func (self Conc) All() error { return self.RunAll().Err() }

/*
Same as `cc.Conc.All` but instead of returning nil or non-nil `error`, always
returns `cc.Errs`, which may or may not be empty. Should be used ONLY if you
know what to do with the error slice. Otherwise, use `cc.Conc.All`.
*/
func (self Conc) RunAll() Errs {
	if self == nil {
		return nil
	}

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

func concRun(fun func(), wg *sync.WaitGroup, ptr *error) {
	defer concDone(wg, ptr)
	fun()
}

func concDone(wg *sync.WaitGroup, ptr *error) {
	wg.Add(-1)

	val := recover()
	if val == nil {
		return
	}

	err, _ := val.(error)
	if err != nil {
		*ptr = err
		return
	}

	panic(val)
}
