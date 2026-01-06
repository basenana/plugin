package utils

import (
	"testing"
)

func TestMarshalMap(t *testing.T) {
	type Test struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	obj := Test{Name: "Alice", Age: 30}
	result := MarshalMap(obj)

	if result == nil {
		t.Fatal("MarshalMap returned nil")
	}

	if result["name"] != "Alice" {
		t.Errorf("expected name 'Alice', got '%v'", result["name"])
	}

	if result["age"] != float64(30) {
		t.Errorf("expected age 30, got '%v'", result["age"])
	}
}

func TestMarshalMapNil(t *testing.T) {
	result := MarshalMap(nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got '%v'", result)
	}
}

func TestUnmarshalMap(t *testing.T) {
	type Test struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := map[string]interface{}{
		"name": "Bob",
		"age":  float64(25),
	}

	var obj Test
	UnmarshalMap(data, &obj)

	if obj.Name != "Bob" {
		t.Errorf("expected name 'Bob', got '%s'", obj.Name)
	}

	if obj.Age != 25 {
		t.Errorf("expected age 25, got %d", obj.Age)
	}
}

func TestUnmarshalMapNil(t *testing.T) {
	var obj struct{ Name string }
	UnmarshalMap(nil, &obj)
	if obj.Name != "" {
		t.Errorf("expected empty name, got '%s'", obj.Name)
	}
}

func TestRoundTrip(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		City string `json:"city"`
	}

	original := Person{Name: "Charlie", City: "NYC"}
	mapped := MarshalMap(original)

	var restored Person
	UnmarshalMap(mapped, &restored)

	if original.Name != restored.Name {
		t.Errorf("name mismatch: got '%s', expected '%s'", restored.Name, original.Name)
	}

	if original.City != restored.City {
		t.Errorf("city mismatch: got '%s', expected '%s'", restored.City, original.City)
	}
}
