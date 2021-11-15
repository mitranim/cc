package cc_test

import (
	"errors"
	"fmt"
	"io"
	r "reflect"
	"testing"

	"github.com/mitranim/cc"
)

var (
	testErr0 = fmt.Errorf(`test err 0`)
	testErr1 = fmt.Errorf(`test err 1`)
	testErr2 = fmt.Errorf(`test err 2`)

	testErrA = Err(`test err A`)
	testErrB = Err(`test err B`)
)

type Err string

func (self Err) Error() string { return string(self) }

func TestErrs_Error(t *testing.T) {
	eq(t, ``, cc.Errs(nil).Error())
	eq(t, ``, cc.Errs{}.Error())
	eq(t, `test err 0`, cc.Errs{nil, testErr0, nil}.Error())
	eq(t, `[cc] multiple errors; test err 0; test err 1`, cc.Errs{nil, testErr0, nil, testErr1, nil}.Error())
}

func TestErrs_Unwrap(t *testing.T) {
	eq(t, nil, cc.Errs(nil).Unwrap())
	eq(t, nil, cc.Errs{}.Unwrap())
	eq(t, testErr0, cc.Errs{nil, testErr0}.Unwrap())
	eq(t, testErr0, cc.Errs{nil, testErr0, testErr1}.Unwrap())
}

func TestErrs_Is(t *testing.T) {
	eq(t, false, errors.Is(cc.Errs(nil), io.EOF))
	eq(t, false, errors.Is(cc.Errs(nil), testErr0))

	eq(t, false, errors.Is(cc.Errs{nil, testErr0, nil, testErr1, nil}, io.EOF))
	eq(t, false, errors.Is(cc.Errs{nil, testErr0, nil, testErr1, nil}, testErr2))

	eq(t, true, errors.Is(cc.Errs{nil, testErr0, nil, testErr1, nil}, testErr0))
	eq(t, true, errors.Is(cc.Errs{nil, testErr0, nil, testErr1, nil}, testErr1))

	eq(t, true, errors.Is(cc.Errs{nil, fmt.Errorf(`%w`, testErr0), nil, testErr1, nil}, testErr0))
	eq(t, true, errors.Is(cc.Errs{nil, fmt.Errorf(`%w`, io.EOF), nil, testErr1, nil}, io.EOF))

	eq(t, true, errors.Is(cc.Errs{nil, testErr0, nil, fmt.Errorf(`%w`, testErr1), nil}, testErr1))
	eq(t, true, errors.Is(cc.Errs{nil, testErr0, nil, fmt.Errorf(`%w`, io.EOF), nil}, io.EOF))
}

func TestErrs_As(t *testing.T) {
	test := func(ok bool, exp Err, src cc.Errs) {
		t.Helper()

		var tar Err
		eq(t, ok, errors.As(src, &tar))
		eq(t, exp, tar)
	}

	test(false, ``, cc.Errs(nil))
	test(false, ``, cc.Errs{})
	test(false, ``, cc.Errs{testErr0})
	test(false, ``, cc.Errs{testErr0, testErr1})
	test(true, testErrA, cc.Errs{testErrA, testErr0, testErr1})
	test(true, testErrA, cc.Errs{testErr0, testErrA, testErr1})
	test(true, testErrA, cc.Errs{testErr0, testErr1, testErrA})
	test(true, testErrA, cc.Errs{nil, testErr0, nil, testErr1, nil, testErrA, nil})
	test(true, testErrA, cc.Errs{nil, testErrA, nil, testErrB, nil})
}

func TestErrs_Err(t *testing.T) {
	testEmpty := func(src cc.Errs) {
		t.Helper()
		eq(t, nil, src.Err())
	}

	testEmpty(cc.Errs(nil))
	testEmpty(cc.Errs{})
	testEmpty(cc.Errs{nil, nil, nil})

	testOne := func(exp error) {
		t.Helper()
		eq(t, exp, cc.Errs{nil, exp, nil}.Err())
		eq(t, exp, cc.Errs{exp, nil}.Err())
		eq(t, exp, cc.Errs{nil, exp}.Err())
	}

	testOne(testErr0)
	testOne(testErr1)

	errs := cc.Errs{nil, testErr0, nil, testErr1, nil}
	eq(t, errs, errs.Err())
}

func TestConc(t *testing.T) {
	t.Run(`success`, func(t *testing.T) {
		eq(t, nil, cc.All())
		eq(t, nil, cc.All(nil, nil, nil))
		eq(t, nil, cc.All(func() {}))
		eq(t, nil, cc.All(func() {}, func() {}))
		eq(t, nil, cc.All(func() {}, nil, func() {}))
		eq(t, nil, cc.All(nil, func() {}, nil, func() {}, nil))
	})

	t.Run(`1 panic`, func(t *testing.T) {
		eq(
			t,
			testErr0,
			cc.All(func() { panic(testErr0) }),
		)
	})

	t.Run(`mixed`, func(t *testing.T) {
		eq(
			t,
			cc.Errs{
				nil,
				testErr0,
				nil,
				testErr1,
				nil,
			},
			cc.All(
				func() {},
				func() { panic(testErr0) },
				func() {},
				func() { panic(testErr1) },
				func() {},
			),
		)
	})
}

func BenchmarkAll_one(b *testing.B) {
	for range iter(b.N) {
		benchAllOne()
	}
}

func benchAllOne() {
	_ = cc.All(func() { panic(testErr0) })
}

func BenchmarkConcAll_one(b *testing.B) {
	for range iter(b.N) {
		benchConcAllOne()
	}
}

func benchConcAllOne() {
	conc := make(cc.Conc, 0, 1)
	conc.Add(func() { panic(testErr0) })
	_ = conc.All()
}

func BenchmarkAll(b *testing.B) {
	for range iter(b.N) {
		benchAll()
	}
}

func benchAll() {
	_ = cc.All(
		func() {},
		func() { panic(testErr0) },
		func() {},
		func() { panic(testErr1) },
		func() {},
	)
}

func BenchmarkConc_All(b *testing.B) {
	for range iter(b.N) {
		benchConcAll()
	}
}

func benchConcAll() {
	conc := make(cc.Conc, 0, 5)

	conc.Add(func() {})
	conc.Add(func() { panic(testErr0) })
	conc.Add(func() {})
	conc.Add(func() { panic(testErr1) })
	conc.Add(func() {})

	_ = conc.All()
}

func eq(t testing.TB, exp, act interface{}) {
	t.Helper()
	if !r.DeepEqual(exp, act) {
		t.Fatalf(`
expected (detailed):
	%#[1]v
actual (detailed):
	%#[2]v
expected (simple):
	%[1]v
actual (simple):
	%[2]v
`, exp, act)
	}
}

func iter(count int) []struct{} { return make([]struct{}, count) }
