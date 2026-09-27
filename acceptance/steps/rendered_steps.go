package steps

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/masonhuemmer/m365/acceptance/runtime"
)

var reJSONStep = regexp.MustCompile(`^stdout JSON "([^"]+)" is (.+)$`)

// registerRendered adds the steps used by the SDO-563 features.
func registerRendered() {
	runtime.Register("stdout JSON", func(w *runtime.World, text string) error {
		m := reJSONStep.FindStringSubmatch(text)
		if m == nil {
			w.T.Fatalf("bad step: %s", text)
		}
		var want any
		if err := json.Unmarshal([]byte(m[2]), &want); err != nil {
			w.T.Fatalf("expected value is not JSON: %s", m[2])
		}
		var got any
		if err := json.Unmarshal([]byte(strings.TrimSpace(w.Out)), &got); err != nil {
			w.T.Fatalf("stdout is not JSON: %s", w.Out)
		}
		for _, key := range strings.Split(m[1], ".") {
			got = jsonField(got, key)
		}
		if !reflect.DeepEqual(got, want) {
			w.T.Fatalf("%s = %#v, want %#v (stdout %s)", m[1], got, want, w.Out)
		}
		return nil
	})
}

// jsonField steps into an object key or an array index; nil if absent.
func jsonField(v any, key string) any {
	switch x := v.(type) {
	case map[string]any:
		return x[key]
	case []any:
		i, err := strconv.Atoi(key)
		if err != nil || i < 0 || i >= len(x) {
			return nil
		}
		return x[i]
	}
	return nil
}
