package startup

// Register adds the daemon to OS startup
func Register() error {
	return registerOS()
}

// Deregister removes the daemon from OS startup
func Deregister() error {
	return deregisterOS()
}
