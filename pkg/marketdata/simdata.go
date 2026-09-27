package marketdata

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SimdataFile is a permanently stored simulated price series, used to extend
// an asset's history before its real quotes begin. Files live in a simdata
// directory, one CSV per canonical identifier, with self-describing headers:
//
//	# pofo simdata v1
//	# id: IE000KF370H3
//	# name: WisdomTree US Efficient Core (90/60 replication)
//	# method: 0.90×VFINX + 0.60×(VFITX − cash ^IRX) + 0.10×cash, fees 0.20 %/yr
//	# validation: corr=0.98 vs NTSX over 2018-08-02 → 2026-06-11
//	# generated: 2026-06-12
//	date,close
//	2000-01-03,100.000000
//
// A file whose SUBJECT changes definition mid-record carries the dates of those
// changes in an optional "# junctions:" header (comma-separated, ISO dates),
// read back into Series.Junctions. It is the one piece of the format a consumer
// must honour to be correct rather than merely informed: the step into a
// junction is not a market move.
type SimdataFile struct {
	ID         string
	Name       string
	Method     string
	Validation string
	Generated  string
	Junctions  []time.Time
	Points     []Point
}

// ReadSimdata loads the simulated series stored for the canonical id in a
// directory on disk. ok is false when no file exists.
func ReadSimdata(dir, id string) (*Series, bool, error) {
	return ReadSimdataFS(os.DirFS(dir), id)
}

// ReadSimdataFS is ReadSimdata over any fs.FS, typically the datasets
// embedded in the binary or os.DirFS for development overrides. It reads the
// file named after CanonicalID(id) with ReadCSV's parser, under one rule
// stricter than ReadCSV's: a bundled series is a price or a level, so every
// value must be positive. Symbol is id in upper case and Source "simdata".
func ReadSimdataFS(fsys fs.FS, id string) (s *Series, ok bool, err error) {
	path := sanitizeFilename(CanonicalID(id)) + ".csv"
	f, err := fsys.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	s, err = readSingle(f, strings.ToUpper(id), positive)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", path, err)
	}
	s.Source = "simdata"
	if s.Name == "" {
		s.Name = s.Symbol + " (simdata)"
	}
	return s, true, nil
}

// WriteSimdata stores a simulated series in dir, named after the canonical
// id, in the format read back by ReadSimdata.
func WriteSimdata(dir string, sf *SimdataFile) error {
	if sf.ID == "" || len(sf.Points) == 0 {
		return fmt.Errorf("incomplete simdata for %q", sf.ID)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# pofo simdata v1\n")
	fmt.Fprintf(&b, "# id: %s\n", CanonicalID(sf.ID))
	if sf.Name != "" {
		fmt.Fprintf(&b, "# name: %s\n", sf.Name)
	}
	if sf.Method != "" {
		fmt.Fprintf(&b, "# method: %s\n", sf.Method)
	}
	if sf.Validation != "" {
		fmt.Fprintf(&b, "# validation: %s\n", sf.Validation)
	}
	if sf.Generated != "" {
		fmt.Fprintf(&b, "# generated: %s\n", sf.Generated)
	}
	if len(sf.Junctions) > 0 {
		days := make([]string, len(sf.Junctions))
		for i, t := range sf.Junctions {
			days[i] = t.Format("2006-01-02")
		}
		fmt.Fprintf(&b, "# junctions: %s\n", strings.Join(days, ","))
	}
	b.WriteString("date,close\n")
	for _, p := range sf.Points {
		fmt.Fprintf(&b, "%s,%.6f\n", p.Date.Format("2006-01-02"), p.Close)
	}
	path := filepath.Join(dir, sanitizeFilename(CanonicalID(sf.ID))+".csv")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
