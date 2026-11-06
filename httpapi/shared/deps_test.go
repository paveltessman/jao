package shared

import "testing"

func TestNewDepsPanicsWithNill(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("panic didn't happen")
		}
	}()
	NewDeps(nil)
}
