package series

import (
	"errors"
	"testing"
)

var (
	failFmt    = "%s got '%s' but wanted '%s'"
	failErrFmt = "%s got an error value %s, but wanted %s"
)

// func (s *ConfigSuite) TestLastestLtsSeriesFallback(c *gc.C) {
// 	config.ResetCachedLtsSeries()
// 	s.PatchValue(config.DistroLtsSeries, func() (string, error) {
// 		return "", fmt.Errorf("error")
// 	})
// 	c.Assert(series.LatestLts(), gc.Equals, "xenial")
// }

// func (s *ConfigSuite) TestLastestLtsSeries(c *gc.C) {
// 	config.ResetCachedLtsSeries()
// 	s.PatchValue(config.DistroLtsSeries, func() (string, error) {
// 		return "series", nil
// 	})
// 	c.Assert(series.LatestLts(), gc.Equals, "series")
// }

func TestMain(m *testing.M) {
	// Preserve the original distInfoCmd and restore it after these test runs.
	origDistInfoCmd := distInfoCmd
	origLatestLts := latestLtsSeries
	defer func() {
		distInfoCmd = origDistInfoCmd
		latestLtsSeries = origLatestLts
	}()

	// Run the tests.
	m.Run()

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
		{"Test distroReturns value", func() ([]byte, error) { return []byte(""), nil }, "a series", "a series"},
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
