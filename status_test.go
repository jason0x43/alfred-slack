package main

import "testing"

// TestStatusItems tests status.Items
func TestStatusItems(t *testing.T) {
	t.Run("initial", func(t *testing.T) {
		cache.Users = []User{User{}}
		s := StatusCommand{}
		items, err := s.Items("", "")
		if err != nil {
			t.Fatal("Error getting items:", err)
		}
		if len(items) == 0 {
			t.Fatal("Items should not be empty")
		}
	})
}
