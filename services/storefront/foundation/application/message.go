package application

import "time"

// Message is the durable envelope of an explicitly selected publication.
type Message struct {
	ID               string
	Name             string
	Context          string
	Visibility       Visibility
	ContractVersion  uint32
	AggregateKind    string
	AggregateID      string
	AggregateVersion uint64
	CorrelationID    string
	CausationID      string
	OccurredAt       time.Time
	Payload          any
}

type Encoder func(Message) ([]byte, error)
type Decoder func([]byte) (Message, error)
