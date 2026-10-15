# Value objects

A value object is described by its meaningful attributes. It validates its values
and is normally replaced as a whole when those values change. Equality follows
the values rather than a separate lifecycle identity.

## Go: Money

Money uses integer minor units, validates its currency and amount, and exposes
value equality. Its private fields prevent arbitrary mutation by a caller.

### Go

Source: [domain.go](../../../services/storefront/foundation/domain/domain.go).

```go source="../../../services/storefront/foundation/domain/domain.go"
type Money struct {
	minor    int64
	currency string
}

func NewMoney(minor int64, currency string) (Money, error) {
	if minor < 0 || minor > 1_000_000 || !regexp.MustCompile(`^[A-Z]{3}$`).MatchString(currency) {
		return Money{}, Reject("invalid_money", "A non-negative bounded amount and three-letter currency are required")
	}
	return Money{minor, currency}, nil
}
func (m Money) Minor() int64           { return m.minor }
func (m Money) Currency() string       { return m.currency }
func (m Money) Equal(other Money) bool { return m == other }
```

## Python: CollectionCode

CollectionCode is a frozen dataclass. Its generated equality compares the value;
construction validates the format used at handover.

### Python

Source: [pickup.py](../../../services/operations/src/operations/contexts/collection/domain/pickup.py).

```python source="../../../services/operations/src/operations/contexts/collection/domain/pickup.py"
class CollectionCode:
    """Validates the owner collection credential as a value object."""

    value: str

    def __post_init__(self) -> None:
        if not re.fullmatch(r"[A-Z0-9]{6}", self.value):
            raise Rejection(
                "invalid_collection_code", "Collection codes have six uppercase letters or digits"
            )
```

The declaration is preceded by `@dataclass(frozen=True)` in the linked source.

## TypeScript: a corresponding teaching example

The Engagement model currently uses validated primitive terms and immutable
records; it has no Money wrapper. This standalone example shows the same Money
semantics if a TypeScript context needed them. It is illustrative code, separate
from the current implementation.

```ts
type Money = Readonly<{ minor: number; currency: string }>;

function money(minor: number, currency: string): Money {
  if (!Number.isSafeInteger(minor) || minor < 0 || minor > 1_000_000) {
    throw new Error('Invalid amount');
  }
  if (!/^[A-Z]{3}$/.test(currency)) throw new Error('Invalid currency');
  return Object.freeze({ minor, currency });
}

function sameMoney(left: Money, right: Money): boolean {
  return left.minor === right.minor && left.currency === right.currency;
}
```

TypeScript's `Readonly` is a compile-time restriction. `Object.freeze` supplies
runtime protection for this flat object. JavaScript object identity with `===`
does not provide value equality for separately constructed Money objects.

## Choose the concept from business meaning

A read-only record containing a grant ID remains an identified business record;
immutability alone does not make it a value object. A MenuOffer freezes published
terms within a MenuEdition and has no independent repository. Match the model
to the equality and lifecycle the business needs.

## Related concepts

[Entities](entities.md) · [Invariants](invariants.md) · [Aggregates](aggregates.md) · [Aggregate hydration](aggregate-hydration.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
