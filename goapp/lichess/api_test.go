package lichess

import "testing"

func TestBuildURLParams(t *testing.T) {
	// with nil
	var params map[string]string = nil

	res := buildURLParams(params)
	if res != "" {
		t.Errorf("expected nil to produce empty params string")
	}

	params = make(map[string]string)
	res = buildURLParams(params)
	if res != "" {
		t.Errorf("expected empty map to produce empty params string")
	}

	params["foo"] = "bar"
	res = buildURLParams(params)
	expected := "foo=bar"
	if res != expected {
		t.Errorf("expected params to be %s, but got %s", expected, res)
	}

	params["foobar"] = "baz"
	res = buildURLParams(params)
	expected = "foo=bar&foobar=baz"
	if res != expected {
		t.Errorf("expected params to be %s, but got %s", expected, res)
	}

}
