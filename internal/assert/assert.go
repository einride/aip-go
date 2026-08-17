package assert

import (
	"errors"
	"reflect"
	"testing"
)

type ErrorFunc = func(t *testing.T, err error)

func Equal[T any](t *testing.T, want, got T) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Error(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("want an error, got nil")
	}
}

func ErrorIs(t *testing.T, err, target error) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Fatalf("want error %v, got %v", target, err)
	}
}

func ErrorAs(t *testing.T, err error, target any) {
	t.Helper()
	if !errors.As(err, target) {
		t.Fatalf("want error assignable to %T, got %v", target, err)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Fatal("want true, got false")
	}
}

func False(t *testing.T, got bool) {
	t.Helper()
	if got {
		t.Fatal("want false, got true")
	}
}
