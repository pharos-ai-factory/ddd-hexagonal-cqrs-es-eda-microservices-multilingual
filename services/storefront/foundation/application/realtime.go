package application

// RealtimeEncoder maps an owned state snapshot to a separate browser contract.
// It is composed at the outer boundary; the aggregate knows no client transport.
type RealtimeEncoder func(eventID, owner, kind, aggregateID string, revision uint64, snapshot []byte) ([]byte, error)
