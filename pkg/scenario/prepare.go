package scenario

// Prepare hoists a source's per-draw setup out of the caller's draw loop.
//
// The data-driven sources collapse their Panel into one weighted history
// before sampling from it (Panel.Combine), work that does not depend on the
// rng and that a Monte-Carlo driver would otherwise redo at every path: on a
// four-asset, forty-year monthly panel it costs more than the draw itself.
// Prepare returns an equivalent Source that does it once. The rng is consumed
// in exactly the same order, so a prepared source draws byte-identical paths.
//
// Sources with nothing to hoist come back unchanged, so a driver can always
// call Prepare once before its loop (decumul's does). The combined history is
// FROZEN in the returned value: after changing a Panel or its Weights, prepare
// again rather than reusing an older result.
func Prepare(s Source) Source {
	if p, ok := s.(preparer); ok {
		return p.prepare()
	}
	return s
}

// preparer is the optional side of Source: implemented by the sources that
// combine a Panel, and by the wrappers holding one.
type preparer interface {
	// prepare returns an equivalent source with its rng-independent setup
	// already done.
	prepare() Source
}
