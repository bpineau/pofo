// Package xls reads the legacy Excel workbooks that two data sources of this
// repository still serve: an OLE2 compound file (magic D0 CF 11 E0) whose
// "Workbook" stream is a BIFF8 record stream.
//
// It exists because the standard library has no equivalent of archive/zip for
// OLE2 and this toolkit takes no third-party dependency. Only what a
// values-only dump needs is implemented: sector chains, the directory, the mini
// stream, and the cell records that can hold a number or a string. Formatting,
// formulas, charts and everything else are skipped silently, which is exactly
// what a data feed asks for.
//
// The usual call is Sheet, which walks the container and returns the values of
// the first worksheet; Streams and Cells are the two layers underneath, exposed
// for a caller that needs another stream or another sheet. A date arrives as a
// serial number rather than as a date, so SerialDate turns it back into one.
//
// The package lives under cmd/internal because it is a generator's tool: the
// pofo binary embeds the CSVs these generators write and never parses a
// workbook itself.
package xls
