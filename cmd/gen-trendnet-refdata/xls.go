package main

// A minimal reader for the legacy Excel format the source still serves: an
// OLE2 compound file (magic D0 CF 11 E0) whose "Workbook" stream is a BIFF8
// record stream. Nothing else in the toolkit needs it, and the stdlib has no
// equivalent of archive/zip for OLE2, so the two layers live here.
//
// Only what a values-only dump needs is implemented: sector chains, the
// directory, the mini stream, and the five cell records that can hold a number
// or a string. Formatting, formulas, charts and everything else are skipped
// silently, which is exactly what a data feed asks for.

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode/utf16"
)

// oleMagic opens every OLE2 compound file.
var oleMagic = [8]byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}

// maxRegSect is the first reserved sector id: anything at or above it is a
// terminator (end of chain, free, FAT, DIFAT), never a readable sector.
const maxRegSect = 0xFFFFFFFA

// ole2Streams returns the named streams of an OLE2 compound file. The file is
// a little filesystem: a sector-allocation table (FAT) chains fixed-size
// sectors into streams, a directory names them, and streams under a cutoff
// live packed inside a "mini stream" with an allocation table of their own.
func ole2Streams(data []byte) (map[string][]byte, error) {
	if len(data) < 512 || [8]byte(data[:8]) != oleMagic {
		return nil, fmt.Errorf("not an OLE2 compound file")
	}
	u16 := func(off int) int { return int(binary.LittleEndian.Uint16(data[off:])) }
	u32 := func(off int) uint32 { return binary.LittleEndian.Uint32(data[off:]) }

	sectorSize, miniSize := 1<<u16(0x1e), 1<<u16(0x20)
	if sectorSize < 128 || sectorSize > 1<<20 || miniSize < 8 || miniSize > sectorSize {
		return nil, fmt.Errorf("implausible sector sizes %d/%d", sectorSize, miniSize)
	}
	numFAT, dirStart := int(u32(0x2c)), u32(0x30)
	miniCutoff, miniFATStart := u32(0x38), u32(0x3c)
	difatStart, numDIFAT := u32(0x44), int(u32(0x48))

	// Sector s starts right after the 512-byte header, whatever the sector size.
	sector := func(s uint32) []byte {
		if s >= maxRegSect {
			return nil
		}
		off := (int(s) + 1) * sectorSize
		if off < 0 || off+sectorSize > len(data) {
			return nil
		}
		return data[off : off+sectorSize]
	}

	// The DIFAT lists the sectors that hold the FAT: 109 entries in the header,
	// then chained sectors whose last word points at the next one.
	difat := make([]uint32, 0, 109+numDIFAT*sectorSize/4)
	for i := range 109 {
		difat = append(difat, u32(0x4c+4*i))
	}
	for s, n := difatStart, 0; s < maxRegSect && n < numDIFAT; n++ {
		blk := sector(s)
		if blk == nil {
			return nil, fmt.Errorf("DIFAT sector %d outside the file", s)
		}
		for i := 0; i+4 <= len(blk)-4; i += 4 {
			difat = append(difat, binary.LittleEndian.Uint32(blk[i:]))
		}
		s = binary.LittleEndian.Uint32(blk[len(blk)-4:])
	}
	if numFAT > len(difat) {
		return nil, fmt.Errorf("%d FAT sectors announced, %d listed", numFAT, len(difat))
	}

	fat := make([]uint32, 0, numFAT*sectorSize/4)
	for _, fs := range difat[:numFAT] {
		blk := sector(fs)
		if blk == nil {
			return nil, fmt.Errorf("FAT sector %d outside the file", fs)
		}
		for i := 0; i+4 <= len(blk); i += 4 {
			fat = append(fat, binary.LittleEndian.Uint32(blk[i:]))
		}
	}

	// readChain walks the FAT from start and concatenates the sectors, cut to
	// size when the stream declares one (a chain is a whole number of sectors).
	readChain := func(start uint32, size int) ([]byte, error) {
		var out []byte
		for s, n := start, 0; s < maxRegSect; n++ {
			if int(s) >= len(fat) {
				return nil, fmt.Errorf("sector %d outside the allocation table", s)
			}
			if n > len(fat) {
				return nil, fmt.Errorf("cyclic sector chain from %d", start)
			}
			blk := sector(s)
			if blk == nil {
				return nil, fmt.Errorf("sector %d outside the file", s)
			}
			out = append(out, blk...)
			s = fat[s]
		}
		if size > 0 && size < len(out) {
			out = out[:size]
		}
		return out, nil
	}

	dirData, err := readChain(dirStart, 0)
	if err != nil {
		return nil, fmt.Errorf("directory: %w", err)
	}
	type entry struct {
		name  string
		kind  byte
		start uint32
		size  int
	}
	var entries []entry
	for i := 0; i+128 <= len(dirData); i += 128 {
		e := dirData[i : i+128]
		nameLen := int(binary.LittleEndian.Uint16(e[0x40:]))
		entries = append(entries, entry{
			name:  utf16Name(e[:max(0, min(nameLen-2, 0x40))]),
			kind:  e[0x42],
			start: binary.LittleEndian.Uint32(e[0x74:]),
			size:  int(binary.LittleEndian.Uint32(e[0x78:])),
		})
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("empty directory")
	}

	// The root entry owns the mini stream; the mini FAT chains it like the FAT
	// chains sectors, only in miniSize pieces.
	var miniStream, miniFATRaw []byte
	if root := entries[0]; root.start < maxRegSect {
		if miniStream, err = readChain(root.start, root.size); err != nil {
			return nil, fmt.Errorf("mini stream: %w", err)
		}
	}
	if miniFATStart < maxRegSect {
		if miniFATRaw, err = readChain(miniFATStart, 0); err != nil {
			return nil, fmt.Errorf("mini FAT: %w", err)
		}
	}
	readMini := func(start uint32, size int) []byte {
		out := make([]byte, 0, size)
		for s := start; s < maxRegSect && len(out) < size; {
			off := int(s) * miniSize
			if off+miniSize > len(miniStream) || int(s)*4+4 > len(miniFATRaw) {
				break
			}
			out = append(out, miniStream[off:off+miniSize]...)
			s = binary.LittleEndian.Uint32(miniFATRaw[int(s)*4:])
		}
		if size < len(out) {
			out = out[:size]
		}
		return out
	}

	const kindStream = 2
	streams := make(map[string][]byte)
	for _, e := range entries {
		if e.kind != kindStream || e.size == 0 {
			continue
		}
		if uint32(e.size) < miniCutoff {
			streams[e.name] = readMini(e.start, e.size)
			continue
		}
		b, err := readChain(e.start, e.size)
		if err != nil {
			return nil, fmt.Errorf("stream %q: %w", e.name, err)
		}
		streams[e.name] = b
	}
	return streams, nil
}

// utf16Name decodes a directory entry's UTF-16LE name.
func utf16Name(b []byte) string {
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+2 <= len(b); i += 2 {
		u = append(u, binary.LittleEndian.Uint16(b[i:]))
	}
	return string(utf16.Decode(u))
}

// BIFF8 record ids. A record is a 2-byte id, a 2-byte length and that many
// bytes of body; a record too long for 8224 bytes spills into CONTINUE records.
const (
	recFormula   = 0x0006
	recContinue  = 0x003C
	recMULRK     = 0x00BD
	recSST       = 0x00FC
	recLABELSST  = 0x00FD
	recNUMBER    = 0x0203
	recRK        = 0x027E
	recBOF       = 0x0809
	substreamWKS = 0x0010 // a BOF that opens a worksheet, not the globals
)

// cell is one value of the sheet, either a number or a string.
type cell struct {
	row, col int
	num      float64
	text     string
	isText   bool
}

type record struct {
	id   uint16
	body []byte
}

// biffCells returns the values of the workbook's FIRST worksheet, in the order
// the file stores them. Cells the sheet leaves empty are simply absent.
func biffCells(wb []byte) []cell {
	var recs []record
	for pos := 0; pos+4 <= len(wb); {
		id := binary.LittleEndian.Uint16(wb[pos:])
		n := int(binary.LittleEndian.Uint16(wb[pos+2:]))
		if pos+4+n > len(wb) {
			break // truncated tail: keep what parsed
		}
		recs = append(recs, record{id, wb[pos+4 : pos+4+n]})
		pos += 4 + n
	}
	sst := sharedStrings(recs)

	sheet := -1
	var out []cell
	for _, r := range recs {
		if r.id == recBOF {
			if len(r.body) >= 4 && binary.LittleEndian.Uint16(r.body[2:]) == substreamWKS {
				sheet++
			}
			continue
		}
		if sheet != 0 {
			continue
		}
		switch {
		case r.id == recLABELSST && len(r.body) >= 10:
			row, col := int(binary.LittleEndian.Uint16(r.body)), int(binary.LittleEndian.Uint16(r.body[2:]))
			if i := int(binary.LittleEndian.Uint32(r.body[6:])); i < len(sst) {
				out = append(out, cell{row: row, col: col, text: sst[i], isText: true})
			}
		case r.id == recNUMBER && len(r.body) >= 14:
			row, col := int(binary.LittleEndian.Uint16(r.body)), int(binary.LittleEndian.Uint16(r.body[2:]))
			out = append(out, cell{row: row, col: col, num: math.Float64frombits(binary.LittleEndian.Uint64(r.body[6:]))})
		case r.id == recRK && len(r.body) >= 10:
			row, col := int(binary.LittleEndian.Uint16(r.body)), int(binary.LittleEndian.Uint16(r.body[2:]))
			out = append(out, cell{row: row, col: col, num: rkValue(binary.LittleEndian.Uint32(r.body[6:]))})
		case r.id == recMULRK && len(r.body) >= 6:
			row, col := int(binary.LittleEndian.Uint16(r.body)), int(binary.LittleEndian.Uint16(r.body[2:]))
			for k := 0; 4+k*6+6 <= len(r.body)-2; k++ {
				rk := binary.LittleEndian.Uint32(r.body[4+k*6+2:])
				out = append(out, cell{row: row, col: col + k, num: rkValue(rk)})
			}
		case r.id == recFormula && len(r.body) >= 20:
			// Only the cached result is kept, and only when it is a number:
			// 0xFFFF in the two high bytes flags a string, error or blank.
			if binary.LittleEndian.Uint16(r.body[12:]) == 0xFFFF {
				continue
			}
			row, col := int(binary.LittleEndian.Uint16(r.body)), int(binary.LittleEndian.Uint16(r.body[2:]))
			out = append(out, cell{row: row, col: col, num: math.Float64frombits(binary.LittleEndian.Uint64(r.body[6:]))})
		}
	}
	return out
}

// rkValue decodes the packed RK number: the top 30 bits are either a 30-bit
// signed integer or the high half of an IEEE double, and the low bit says the
// result was multiplied by 100 to fit.
func rkValue(rk uint32) float64 {
	v := rk &^ 3
	var f float64
	if rk&2 != 0 {
		f = float64(int32(v) >> 2)
	} else {
		f = math.Float64frombits(uint64(v) << 32)
	}
	if rk&1 != 0 {
		f /= 100
	}
	return f
}

// sharedStrings decodes the workbook's shared string table (an SST record and
// the CONTINUE records that follow it), which is where every string cell of a
// BIFF8 file actually lives.
func sharedStrings(recs []record) []string {
	var r *sstReader
	unique := 0
	for i, rec := range recs {
		if rec.id != recSST || len(rec.body) < 8 {
			continue
		}
		unique = int(binary.LittleEndian.Uint32(rec.body[4:]))
		r = &sstReader{buf: rec.body[8:]}
		for j := i + 1; j < len(recs) && recs[j].id == recContinue; j++ {
			r.conts = append(r.conts, recs[j].body)
		}
		break
	}
	if r == nil {
		return nil
	}
	out := make([]string, 0, unique)
	for range unique {
		s, ok := r.next()
		if !ok {
			break
		}
		out = append(out, s)
	}
	return out
}

// sstReader walks the shared string table across its CONTINUE boundaries. A
// string may be cut in the middle, and the continuation then restates whether
// its remainder is 16-bit or 8-bit, which is the whole subtlety of the format.
type sstReader struct {
	buf   []byte
	conts [][]byte
	ci    int
}

// next reads one string: a character count, a flags byte, the optional
// rich-text and phonetic sizes, the characters themselves, then those two
// blocks skipped over.
func (r *sstReader) next() (string, bool) {
	for len(r.buf) < 3 && r.ci < len(r.conts) {
		r.buf = append(r.buf, r.conts[r.ci]...)
		r.ci++
	}
	if len(r.buf) < 3 {
		return "", false
	}
	n := int(binary.LittleEndian.Uint16(r.buf))
	flags := r.buf[2]
	r.buf = r.buf[3:]
	var rich, phonetic int
	if flags&8 != 0 {
		if len(r.buf) < 2 {
			return "", false
		}
		rich, r.buf = int(binary.LittleEndian.Uint16(r.buf)), r.buf[2:]
	}
	if flags&4 != 0 {
		if len(r.buf) < 4 {
			return "", false
		}
		phonetic, r.buf = int(binary.LittleEndian.Uint32(r.buf)), r.buf[4:]
	}

	wide := flags&1 != 0
	var sb strings.Builder
	for n > 0 {
		w := 1
		if wide {
			w = 2
		}
		take := min(len(r.buf)/w, n)
		if take > 0 {
			sb.WriteString(decodeChars(r.buf[:take*w], wide))
			r.buf = r.buf[take*w:]
			n -= take
		}
		if n > 0 {
			if r.ci >= len(r.conts) || len(r.conts[r.ci]) == 0 {
				return sb.String(), false
			}
			r.buf = r.conts[r.ci]
			r.ci++
			wide = r.buf[0]&1 != 0
			r.buf = r.buf[1:]
		}
	}
	for skip := rich*4 + phonetic; skip > 0; {
		if len(r.buf) >= skip {
			r.buf = r.buf[skip:]
			break
		}
		skip -= len(r.buf)
		if r.ci >= len(r.conts) {
			r.buf = nil
			break
		}
		r.buf = r.conts[r.ci]
		r.ci++
	}
	return sb.String(), true
}

// decodeChars turns a run of BIFF characters into a string: UTF-16LE when the
// string is wide, one byte per character (Latin-1) when it is compressed.
func decodeChars(b []byte, wide bool) string {
	if !wide {
		rs := make([]rune, len(b))
		for i, c := range b {
			rs[i] = rune(c)
		}
		return string(rs)
	}
	u := make([]uint16, 0, len(b)/2)
	for i := 0; i+2 <= len(b); i += 2 {
		u = append(u, binary.LittleEndian.Uint16(b[i:]))
	}
	return string(utf16.Decode(u))
}
