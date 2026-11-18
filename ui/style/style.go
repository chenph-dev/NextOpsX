// Package style holds the UI scale shared by the view packages.
package style

// uiScale is the user's interface scale (1 = 100%). It is set on the UI
// thread each frame by the shell's theme setup.
var uiScale float32 = 1

// SetScale stores the interface scale. A non-positive value resets it to 1.
func SetScale(v float32) {
	if v <= 0 {
		v = 1
	}
	uiScale = v
}

// FontSize scales an explicit text size by the interface scale.
func FontSize(v float32) float32 { return v * uiScale }
