package workspace

// IsSQLProtocol reports whether a protocol uses the SQL workbench. Redis
// and Elasticsearch are database connectors too, but have their own panels.
func (a *State) IsSQLProtocol(protocol string) bool {
	return a.IsDBProtocol(protocol) &&
		!a.IsRedisProtocol(protocol) && !a.IsESProtocol(protocol)
}
