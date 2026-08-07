package xls

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
	"unicode/utf16"
)

// The sources are fetched at generation time only, so the reader is tested on
// a workbook this file builds itself: a real OLE2 container (regular chain for
// the big stream, mini stream for the small one) holding a BIFF8 record stream
// that exercises every cell record the reader understands.

func TestSheetReadsEveryCellRecord(t *testing.T) {
	cells, err := Sheet(craftedWorkbook(t))
	if err != nil {
		t.Fatalf("Sheet: %v", err)
	}
	want := []Cell{
		{Row: 0, Col: 5, Text: header, IsText: true}, // LABELSST
		{Row: 2, Col: 0, Num: 1987},                  // NUMBER
		{Row: 2, Col: 1, Num: 0.11},                  // MULRK, three cells in one record
		{Row: 2, Col: 2, Num: -0.02},
		{Row: 2, Col: 3, Num: 0.03},
		{Row: 3, Col: 1, Num: 0.05}, // RK
		{Row: 3, Col: 2, Num: -0.07},
	}
	if len(cells) != len(want) {
		t.Fatalf("got %d cells, want %d: %v", len(cells), len(want), cells)
	}
	for i, w := range want {
		got := cells[i]
		if got.Row != w.Row || got.Col != w.Col || got.IsText != w.IsText ||
			got.Text != w.Text || math.Abs(got.Num-w.Num) > 1e-12 {
			t.Errorf("cell %d = %+v, want %+v", i, got, w)
		}
	}
}

// The small stream lives in the mini stream, the big one in a regular sector
// chain: the two paths of the container layer.
func TestStreams(t *testing.T) {
	streams, err := Streams(craftedWorkbook(t))
	if err != nil {
		t.Fatalf("Streams: %v", err)
	}
	if got := len(streams["Workbook"]); got < 4096 {
		t.Errorf("Workbook stream is %d bytes, want the regular-chain path", got)
	}
	summary, ok := streams["Summary"]
	if !ok {
		t.Fatalf("no Summary stream, got %v", keys(streams))
	}
	if len(summary) != len(miniPayload) || string(summary) != string(miniPayload) {
		t.Errorf("Summary stream = %d bytes, want the %d written", len(summary), len(miniPayload))
	}
}

func TestStreamsRejectsForeignFormat(t *testing.T) {
	if _, err := Streams([]byte("PK\x03\x04 an OOXML zip, not a compound file")); err == nil {
		t.Fatal("a zip parsed as an OLE2 container")
	}
	if _, err := Sheet([]byte("PK\x03\x04 an OOXML zip, not a compound file")); err == nil {
		t.Fatal("Sheet accepted a zip")
	}
}

// The serial numbers are the ones the sources actually serve, checked against
// the dates their CSV copies carry.
func TestSerialDate(t *testing.T) {
	for _, c := range []struct {
		serial float64
		want   string
	}{
		{36528, "2000-01-03"}, // the first day of the pure-trend daily dump
		{36529, "2000-01-04"},
		{46240, "2026-08-06"},
		{1, "1899-12-31"}, // the epoch is two days before 1900-01-01
	} {
		if got := SerialDate(c.serial).Format("2006-01-02"); got != c.want {
			t.Errorf("SerialDate(%.0f) = %s, want %s", c.serial, got, c.want)
		}
		if d := SerialDate(c.serial); d.Location() != time.UTC {
			t.Errorf("SerialDate(%.0f) is in %s, want UTC", c.serial, d.Location())
		}
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// header is the string cell the fixture stores, through the shared string
// table that every BIFF8 string cell goes through.
const header = "TEST Index Historical Data ROR"

// miniPayload is the small stream stored in the container's mini stream.
var miniPayload = []byte("a stream under the cutoff, packed into the mini stream")

// craftedWorkbook builds the whole fixture: the BIFF records first, then the
// OLE2 container around them.
func craftedWorkbook(t *testing.T) []byte {
	t.Helper()
	var wb []byte
	rec := func(id uint16, body []byte) {
		wb = binary.LittleEndian.AppendUint16(wb, id)
		wb = binary.LittleEndian.AppendUint16(wb, uint16(len(body)))
		wb = append(wb, body...)
	}
	bof := func(substream uint16) []byte {
		b := make([]byte, 16)
		binary.LittleEndian.PutUint16(b, 0x0600) // BIFF8
		binary.LittleEndian.PutUint16(b[2:], substream)
		return b
	}
	// Workbook globals: the shared string table lives here.
	rec(recBOF, bof(0x0005))
	sst := make([]byte, 8)
	binary.LittleEndian.PutUint32(sst, 1)     // total references
	binary.LittleEndian.PutUint32(sst[4:], 1) // unique strings
	sst = binary.LittleEndian.AppendUint16(sst, uint16(len(header)))
	sst = append(sst, 0) // flags: compressed, no rich text, no phonetics
	sst = append(sst, []byte(header)...)
	rec(recSST, sst)

	// The worksheet substream.
	rec(recBOF, bof(substreamWKS))
	rec(recLABELSST, cellHead(0, 5, make([]byte, 4))) // string index 0
	rec(recNUMBER, cellHead(2, 0, f64(1987)))
	rec(recMULRK, mulrk(2, 1, []float64{0.11, -0.02, 0.03}))
	rec(recRK, cellHead(3, 1, rk(0.05)))
	rec(recRK, cellHead(3, 2, rk(-0.07)))
	rec(0x000A, nil) // EOF

	// Padding, so the stream needs the regular sector chain rather than the
	// mini stream: exactly what the real files do.
	rec(0x005D, make([]byte, 4096))

	return ole2Container(t, wb, miniPayload)
}

// cellHead prefixes a cell record's row, column and format index.
func cellHead(row, col int, rest []byte) []byte {
	b := make([]byte, 6)
	binary.LittleEndian.PutUint16(b, uint16(row))
	binary.LittleEndian.PutUint16(b[2:], uint16(col))
	return append(b, rest...)
}

func f64(v float64) []byte {
	return binary.LittleEndian.AppendUint64(nil, math.Float64bits(v))
}

// rk packs a value as an RK number in its "hundredths of an integer" form,
// which is exact for a return quoted in whole percent.
func rk(v float64) []byte {
	n := int32(math.Round(v * 100))
	return binary.LittleEndian.AppendUint32(nil, uint32(n<<2)|3)
}

// mulrk builds one MULRK record: a run of RK cells on the same row.
func mulrk(row, col0 int, vals []float64) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint16(b, uint16(row))
	binary.LittleEndian.PutUint16(b[2:], uint16(col0))
	for _, v := range vals {
		b = binary.LittleEndian.AppendUint16(b, 0) // format index
		b = append(b, rk(v)...)
	}
	return binary.LittleEndian.AppendUint16(b, uint16(col0+len(vals)-1))
}

// ole2Container wraps two streams in a compound file: "Workbook" in a regular
// sector chain and "Summary" in the mini stream.
//
// Layout, 512-byte sectors: 0 the FAT, 1 the directory, 2 the mini FAT, 3 the
// mini stream, 4 onwards the workbook.
func ole2Container(t *testing.T, workbook, small []byte) []byte {
	t.Helper()
	const (
		sectorSize = 512
		miniSize   = 64
		endOfChain = 0xFFFFFFFE
		freeSect   = 0xFFFFFFFF
		fatSect    = 0xFFFFFFFD
	)
	pad := func(b []byte, to int) []byte {
		for len(b)%to != 0 {
			b = append(b, 0)
		}
		return b
	}
	wbSectors := (len(workbook) + sectorSize - 1) / sectorSize
	if wbSectors < 2 {
		t.Fatalf("fixture workbook is only %d bytes: it would not use the regular chain", len(workbook))
	}

	fat := make([]uint32, sectorSize/4)
	for i := range fat {
		fat[i] = freeSect
	}
	fat[0], fat[1], fat[2], fat[3] = fatSect, endOfChain, endOfChain, endOfChain
	for i := range wbSectors {
		fat[4+i] = uint32(5 + i)
	}
	fat[4+wbSectors-1] = endOfChain

	miniFAT := make([]uint32, sectorSize/4)
	for i := range miniFAT {
		miniFAT[i] = freeSect
	}
	for i := range (len(small) + miniSize - 1) / miniSize {
		miniFAT[i] = uint32(i + 1)
	}
	miniFAT[(len(small)+miniSize-1)/miniSize-1] = endOfChain

	dirEntry := func(name string, kind byte, start uint32, size int) []byte {
		e := make([]byte, 128)
		u := utf16.Encode([]rune(name + "\x00"))
		for i, c := range u {
			binary.LittleEndian.PutUint16(e[2*i:], c)
		}
		binary.LittleEndian.PutUint16(e[0x40:], uint16(2*len(u)))
		e[0x42] = kind
		binary.LittleEndian.PutUint32(e[0x74:], start)
		binary.LittleEndian.PutUint32(e[0x78:], uint32(size))
		return e
	}
	var dir []byte
	dir = append(dir, dirEntry("Root Entry", 5, 3, sectorSize)...)
	dir = append(dir, dirEntry("Workbook", 2, 4, len(workbook))...)
	dir = append(dir, dirEntry("Summary", 2, 0, len(small))...)
	dir = append(dir, make([]byte, 128)...) // one unused slot
	dir = pad(dir, sectorSize)

	head := make([]byte, sectorSize)
	copy(head, oleMagic[:])
	binary.LittleEndian.PutUint16(head[0x1e:], 9) // 512-byte sectors
	binary.LittleEndian.PutUint16(head[0x20:], 6) // 64-byte mini sectors
	binary.LittleEndian.PutUint32(head[0x2c:], 1) // one FAT sector
	binary.LittleEndian.PutUint32(head[0x30:], 1) // directory at sector 1
	binary.LittleEndian.PutUint32(head[0x38:], 4096)
	binary.LittleEndian.PutUint32(head[0x3c:], 2) // mini FAT at sector 2
	binary.LittleEndian.PutUint32(head[0x44:], endOfChain)
	binary.LittleEndian.PutUint32(head[0x48:], 0)
	for i := range 109 {
		binary.LittleEndian.PutUint32(head[0x4c+4*i:], freeSect)
	}
	binary.LittleEndian.PutUint32(head[0x4c:], 0) // the FAT is in sector 0

	out := head
	for _, w := range fat {
		out = binary.LittleEndian.AppendUint32(out, w)
	}
	out = append(out, dir...)
	for _, w := range miniFAT {
		out = binary.LittleEndian.AppendUint32(out, w)
	}
	out = pad(append(out, small...), sectorSize)
	return pad(append(out, workbook...), sectorSize)
}
