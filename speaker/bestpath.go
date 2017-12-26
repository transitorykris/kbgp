package speaker

// DefaultBestPathSelection implements the best path selection as defined
// in RFC 4271
type DefaultBestPathSelection struct{}

// Compare implements BestPathSelecter
func (d *DefaultBestPathSelection) Compare(nlris ...NLRI) NLRI {
	// TODO: Implement me
	return NLRI{}
}
