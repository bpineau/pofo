package marketdata

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownIdentifier reports that no source quotes an identifier and that
// none of them failed to answer: the identifier itself is the problem (a
// typo, an invented ticker, a line no source carries), so the same request
// will not start working later. Detect it with errors.Is, or with errors.As
// on *UnknownIdentifierError to recover the identifier itself.
//
// A fetch that failed because a source could not be reached (a network
// error, a rate limit, an HTTP 5xx, an unreadable payload) is deliberately
// NOT marked: such a failure says nothing about the identifier, and a caller
// must retry rather than blame the user.
var ErrUnknownIdentifier = errors.New("no source quotes this identifier")

// UnknownIdentifierError names the identifier no source quotes, so a caller
// can report it without parsing a message. Its own text is the per-source
// failure summary, which it also wraps, so a log line loses nothing.
type UnknownIdentifierError struct {
	ID       string // the canonical identifier that resolved to nothing
	Failures error  // what every source answered, for the log
}

func (e *UnknownIdentifierError) Error() string { return e.Failures.Error() }

func (e *UnknownIdentifierError) Unwrap() error { return e.Failures }

// Is makes errors.Is(err, ErrUnknownIdentifier) true for this error, the
// idiom *fs.PathError uses for fs.ErrNotExist.
func (e *UnknownIdentifierError) Is(target error) bool { return target == ErrUnknownIdentifier }

// errAbsent marks a fetch failure that is EVIDENCE ABOUT THE IDENTIFIER: the
// source answered (or could never be consulted for that symbol at all) and
// holds nothing under that name. Its absence from an error chain therefore
// means "a source did not answer", which is the safe default: an
// unclassified failure reads as an outage, so a new source or a new error
// path can only ever cost a friendlier status, never blame a visitor for
// something that was not their fault.
var errAbsent = errors.New("no such instrument at this source")

// marker attaches a sentinel to an error without changing what it says: the
// sentinel is for errors.Is, the message stays the one written for the
// reader.
type marker struct {
	err  error
	mark error
}

func (m marker) Error() string   { return m.err.Error() }
func (m marker) Unwrap() []error { return []error{m.err, m.mark} }

// markAbsent tags err as a source reporting that it holds nothing.
func markAbsent(err error) error { return marker{err: err, mark: errAbsent} }

// absent reports whether err (or anything it wraps) is such a report.
func absent(err error) bool { return errors.Is(err, errAbsent) }

// markAbsentIfAll marks a summary error as an absence only when every
// failure it combines is one: a single unreachable source makes the whole
// summary uninformative about the identifier.
func markAbsentIfAll(err error, causes ...error) error {
	for _, cause := range causes {
		if !absent(cause) {
			return err
		}
	}
	return markAbsent(err)
}

// noUsableSource is the failure summary of an identifier nothing could serve.
// kind is how it was read ("ticker", "ISIN"); absentOnly says whether every
// source reported an absence rather than a failure, and only then, and only
// with something actually reported, does the summary become an
// *UnknownIdentifierError.
func noUsableSource(kind, id string, failures []string, absentOnly bool) error {
	err := fmt.Errorf("%s %s: no usable source (%s)", kind, id, strings.Join(failures, "; "))
	if !absentOnly || len(failures) == 0 {
		return err
	}
	return &UnknownIdentifierError{ID: id, Failures: err}
}
