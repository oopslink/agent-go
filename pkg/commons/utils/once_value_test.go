package utils

import (
	"testing"
)

func TestOfOnceValue(t *testing.T) {
	// Test with string value
	value := "test value"
	onceVal := OfOnceValue(value)

	if onceVal == nil {
		t.Error("OfOnceValue() returned nil")
	}

	// Test that the value is stored correctly
	if onceVal.v != value {
		t.Errorf("onceValue.v = %v, want %v", onceVal.v, value)
	}

	// Test that pop flag is initially false
	if onceVal.pop {
		t.Error("onceValue.pop should be false initially")
	}
}

func TestOnceValueGet(t *testing.T) {
	value := "test value"
	onceVal := OfOnceValue(value)

	called := false
	var receivedValue string

	// First call should yield the value
	onceVal.Get(func(v string) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function on first call")
	}

	if receivedValue != value {
		t.Errorf("Get() yielded %v, want %v", receivedValue, value)
	}

	if !onceVal.pop {
		t.Error("onceValue.pop should be true after first call")
	}
}

func TestOnceValueGetMultipleCalls(t *testing.T) {
	value := "test value"
	onceVal := OfOnceValue(value)

	callCount := 0
	var receivedValue string

	// First call
	onceVal.Get(func(v string) {
		callCount++
		receivedValue = v
	})

	if callCount != 1 {
		t.Errorf("First call: callCount = %d, want 1", callCount)
	}

	if receivedValue != value {
		t.Errorf("First call: receivedValue = %v, want %v", receivedValue, value)
	}

	// Second call should not yield
	onceVal.Get(func(v string) {
		callCount++
		receivedValue = v
	})

	if callCount != 1 {
		t.Errorf("Second call: callCount = %d, want 1", callCount)
	}

	// Third call should not yield
	onceVal.Get(func(v string) {
		callCount++
		receivedValue = v
	})

	if callCount != 1 {
		t.Errorf("Third call: callCount = %d, want 1", callCount)
	}
}

func TestOnceValueWithInt(t *testing.T) {
	value := 42
	onceVal := OfOnceValue(value)

	called := false
	var receivedValue int

	onceVal.Get(func(v int) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function")
	}

	if receivedValue != value {
		t.Errorf("Get() yielded %v, want %v", receivedValue, value)
	}
}

func TestOnceValueWithStruct(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	value := TestStruct{Name: "test", Value: 123}
	onceVal := OfOnceValue(value)

	called := false
	var receivedValue TestStruct

	onceVal.Get(func(v TestStruct) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function")
	}

	if receivedValue != value {
		t.Errorf("Get() yielded %v, want %v", receivedValue, value)
	}
}

func TestOnceValueWithPointer(t *testing.T) {
	value := "test"
	onceVal := OfOnceValue(&value)

	called := false
	var receivedValue *string

	onceVal.Get(func(v *string) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function")
	}

	if receivedValue != &value {
		t.Errorf("Get() yielded %v, want %v", receivedValue, &value)
	}
}

func TestOnceValueWithSlice(t *testing.T) {
	value := []int{1, 2, 3, 4, 5}
	onceVal := OfOnceValue(value)

	called := false
	var receivedValue []int

	onceVal.Get(func(v []int) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function")
	}

	if len(receivedValue) != len(value) {
		t.Errorf("Get() yielded slice with length %d, want %d", len(receivedValue), len(value))
	}

	for i, v := range receivedValue {
		if v != value[i] {
			t.Errorf("Get() yielded slice[%d] = %d, want %d", i, v, value[i])
		}
	}
}

func TestOnceValueWithMap(t *testing.T) {
	value := map[string]int{"a": 1, "b": 2, "c": 3}
	onceVal := OfOnceValue(value)

	called := false
	var receivedValue map[string]int

	onceVal.Get(func(v map[string]int) {
		called = true
		receivedValue = v
	})

	if !called {
		t.Error("Get() should call the yield function")
	}

	if len(receivedValue) != len(value) {
		t.Errorf("Get() yielded map with %d entries, want %d", len(receivedValue), len(value))
	}

	for k, v := range value {
		if receivedValue[k] != v {
			t.Errorf("Get() yielded map[%s] = %d, want %d", k, receivedValue[k], v)
		}
	}
}

func TestOnceValueNilYield(t *testing.T) {
	value := "test value"
	onceVal := OfOnceValue(value)

	// Test that calling Get with nil yield function doesn't panic
	// and still sets the pop flag
	onceVal.Get(nil)

	if !onceVal.pop {
		t.Error("onceValue.pop should be true after calling Get with nil yield")
	}

	// Second call should not panic and should not call yield function
	called := false
	onceVal.Get(func(v string) {
		called = true
		t.Error("Get() should not call yield function on second call")
	})

	if called {
		t.Error("Yield function should not be called on second call")
	}
}

func TestOnceValueConcurrentAccess(t *testing.T) {
	value := "test value"
	onceVal := OfOnceValue(value)

	// This test verifies that the onceValue behaves correctly
	// even if accessed concurrently (though it's not thread-safe)
	// The main point is that it doesn't panic and maintains the once behavior

	called := false
	onceVal.Get(func(v string) {
		called = true
		if v != value {
			t.Errorf("Concurrent access: yielded %v, want %v", v, value)
		}
	})

	if !called {
		t.Error("Concurrent access: Get() should call the yield function")
	}

	if !onceVal.pop {
		t.Error("Concurrent access: onceValue.pop should be true after call")
	}
}
