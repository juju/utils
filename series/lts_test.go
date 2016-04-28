package series

import (
	"errors"
	"testing"
)

var (
	failFmt    = "%s got '%s' but wanted '%s'"
	failErrFmt = "%s got an error value %s, but wanted %s"
)

func TestMain(m *testing.M) {
	// Preserve the original distInfoCmd and restore it after these test runs.
	origDistInfoCmd := distInfoCmd
	defer func() { distInfoCmd = origDistInfoCmd }()

	// Run the tests.
	m.Run()

}

type DummyDefaultSeries struct{}

var dummySeries string
var dummyBool bool

func (_ DummyDefaultSeries) DefaultSeries() (string, bool) { return dummySeries, dummyBool }

// API tests
func TestPreferred(t *testing.T) {
	table := []struct {
		name   string
		dummyS string
		dummyB bool
		want   string
	}{
		{"Test Preferred no default", "", false, "xenial"},
		{"Test Preferred with default", "series", true, "series"},
	}
	d := DummyDefaultSeries{}
	for _, test := range table {
		dummySeries = test.dummyS
		dummyBool = test.dummyB
		got := Preferred(d)
		if got != test.want {
			t.Errorf(failFmt, test.name, got, test.want)
		}
	}

}

func TestLastestLts(t *testing.T) {
	table := []struct {
		name   string
		cmd    func() ([]byte, error)
		latest string
		want   string
	}{
		{"Test latestLtsSeries is set", func() ([]byte, error) { return []byte("nope"), nil }, "latest", "latest"},
		{"Test distroReturns value", func() ([]byte, error) { return []byte("trusty"), nil }, "", "trusty"},
		{"Test fallbackLtsSeries", func() ([]byte, error) { return []byte(""), errors.New("error") }, "", fallbackLtsSeries},
	}
	for _, test := range table {
		latestLtsSeries = test.latest
		distInfoCmd = test.cmd
		got := LatestLts()
		if got != test.want {
			t.Errorf(failFmt, test.name, got, test.want)
		}
	}
}

func TestSetLatestLts(t *testing.T) {
	table := []struct {
		series string
	}{
		{"xenial"},
		{"trusty"},
		{"precise"},
		{"invalid series"},
	}
	orig := latestLtsSeries
	defer func() { latestLtsSeries = orig }()
	for _, test := range table {
		SetLatestLts(test.series)
		got := latestLtsSeries
		if got != test.series {
			t.Errorf(failFmt, "Test SetLatestLts", got, test.series)
		}
	}
}

// Implementation tests

func TestDistroLtsFunc(t *testing.T) {
	table := []struct {
		name   string
		cmd    func() ([]byte, error)
		want   string
		err    bool
		errVal string
	}{
		{"Test valid LTS series", func() ([]byte, error) { return []byte("xenial"), nil }, "xenial", false, ""},
		{"Test another valid LTS series", func() ([]byte, error) { return []byte("trusty"), nil }, "trusty", false, ""},
		{"Test invalid series",
			func() ([]byte, error) { return []byte("foo"), nil },
			"",
			true,
			`not a valid LTS series: "foo"`},
		{"Test invalid LTS series", func() ([]byte, error) { return []byte("wily"), nil }, "", true, `not a valid LTS series: "wily"`},
		{"Test distro-info not found", func() ([]byte, error) {
			return []byte(""), errors.New(`exec: "distro-info": executable file not found in $PATH`)
		}, "", true, `exec: "distro-info": executable file not found in $PATH`},
	}
	for _, test := range table {
		distInfoCmd = test.cmd
		got, gotErr := distroLtsSeriesFunc()
		switch test.err {
		case false:
			if got != test.want {
				t.Errorf(failFmt, test.name, got, test.want)
			}
		case true:
			if gotErr == nil {
				t.Errorf("%s: didn't fail as expected.", test.name)
				t.FailNow()
			}
			if gotErr.Error() != test.errVal {
				t.Errorf(failErrFmt, test.name, gotErr, test.errVal)
			}
		}

	}

}

func TestIsValidLts(t *testing.T) {
	table := []struct {
		value string
		want  bool
	}{
		{"precise", true},
		{"quantal", false},
		{"raring", false},
		{"trusty", true},
		{"wily", false},
		{"xenial", true},
		{"notPresent", false},
	}
	for _, test := range table {
		got := isValidLts(test.value)
		if got != test.want {
			t.Error(failFmt, "isValidLts series", got, test.want)
		}
	}
}
