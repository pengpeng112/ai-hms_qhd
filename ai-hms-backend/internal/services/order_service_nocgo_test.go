//go:build !cgo

package services

import "testing"

func TestOrderServiceRequiresCGO(t *testing.T) {
	t.Skip("order service sqlite-backed tests require cgo-enabled sqlite driver")
}
