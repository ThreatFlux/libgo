package connection

// ConnectionWrapper is an alias type for *libvirtConnection to help with tests.
type ConnectionWrapper = *libvirtConnection

// Cast converts a Connection interface to a ConnectionWrapper when needed.
func Cast(conn Connection) ConnectionWrapper {
	if lc, ok := conn.(*libvirtConnection); ok {
		return lc
	}
	// In tests, we may use a different implementation
	// Return nil as this is only for compile-time compatibility
	return nil
}
