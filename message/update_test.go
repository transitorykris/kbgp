package message

import "testing"

func TestOptionalAttribute(t *testing.T) {
	a := attributeType{flags: optional}
	if !a.optional() {
		t.Error("Expected attribute to be optional")
	}
	if a.wellKnown() {
		t.Error("Did not expect attribute to be well-known")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.optional() {
		t.Error("Expected attribute to be optional")
	}
	if a.wellKnown() {
		t.Error("Did not expect attribute to be well-known")
	}
}

func TestSetOptional(t *testing.T) {
	a := attributeType{flags: wellKnown}
	a.setOptional()
	if a.wellKnown() {
		t.Error("Expected the attribute to be optional")
	}

	a = attributeType{flags: optional}
	a.setOptional()
	if a.wellKnown() {
		t.Error("Expected the attribute to be optional")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestWellKnownAttribute(t *testing.T) {
	a := attributeType{flags: wellKnown}
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	if a.optional() {
		t.Error("Did not expect attribute to be optional")
	}

	a = attributeType{flags: wellKnown | transitive | partial | extendedLength}
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	if a.optional() {
		t.Error("Did not expect attribute to be optional")
	}
}

func TestSetWellKnown(t *testing.T) {
	a := attributeType{flags: optional}
	a.setWellKnown()
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}

	a = attributeType{flags: wellKnown}
	a.setWellKnown()
	if !a.wellKnown() {
		t.Error("Expected attribute to be well-known")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestTransitiveAttribute(t *testing.T) {
	a := attributeType{flags: transitive}
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.nonTransitive() {
		t.Error("Did not expect attribute to be non-transitive")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.nonTransitive() {
		t.Error("Did not expect attribute to be non-transitive")
	}
}

func TestSetTransitive(t *testing.T) {
	a := attributeType{flags: nonTransitive}
	a.setTransitive()
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}

	a = attributeType{flags: transitive}
	a.setTransitive()
	if !a.transitive() {
		t.Error("Expected attribute to be transitive")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestNonTransitiveAttribute(t *testing.T) {
	a := attributeType{flags: nonTransitive}
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}
	if a.transitive() {
		t.Error("Did not expect attribute to be transitive")
	}

	a = attributeType{flags: optional | nonTransitive | partial | extendedLength}
	if !a.nonTransitive() {
		t.Error("Expected attribute to be transitive")
	}
	if a.transitive() {
		t.Error("Did not expect attribute to be transitive")
	}
}

func TestSetNonTransitive(t *testing.T) {
	a := attributeType{flags: transitive}
	a.setNonTransitive()
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}

	a = attributeType{flags: nonTransitive}
	a.setNonTransitive()
	if !a.nonTransitive() {
		t.Error("Expected attribute to be non-transitive")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestPartialAttribute(t *testing.T) {
	a := attributeType{flags: partial}
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	if a.complete() {
		t.Error("Did not expect attribute to be complete")
	}

	a = attributeType{flags: optional | transitive | partial | extendedLength}
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	if a.complete() {
		t.Error("Did not expect attribute to be complete")
	}
}

func TestSetPartial(t *testing.T) {
	a := attributeType{flags: complete}
	a.setPartial()
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}

	a = attributeType{flags: partial}
	a.setPartial()
	if !a.partial() {
		t.Error("Expected attribute to be partial")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestCompleteAttribute(t *testing.T) {
	a := attributeType{flags: complete}
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be partial")
	}

	a = attributeType{flags: optional | transitive | complete | extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be partial")
	}
}

func TestSetComplete(t *testing.T) {
	a := attributeType{flags: partial}
	a.setComplete()
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}

	a = attributeType{flags: complete}
	a.setComplete()
	if !a.complete() {
		t.Error("Expected attribute to be complete")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestExtendedLengthAttribute(t *testing.T) {
	a := attributeType{flags: extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be not extended length")
	}

	a = attributeType{flags: optional | transitive | complete | extendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be not extended length")
	}
}

func TestSetExtendedLength(t *testing.T) {
	a := attributeType{flags: notExtendedLength}
	a.setExtendedLength()
	if !a.extendedLength() {
		t.Error("Expected attribute to be extended length")
	}

	a = attributeType{flags: extendedLength}
	a.setExtendedLength()
	if !a.extendedLength() {
		t.Error("Expected attribute to be extended length")
	}
	// TODO: Should also test that we're not squashing other bits
}

func TestNotExtendedLengthAttribute(t *testing.T) {
	a := attributeType{flags: notExtendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be not extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be  extended length")
	}

	a = attributeType{flags: optional | transitive | complete | notExtendedLength}
	if !a.complete() {
		t.Error("Expected attribute to be not extended length")
	}
	if a.partial() {
		t.Error("Did not expect attribute to be extended length")
	}
}

func TestSetNotExtendedLength(t *testing.T) {
	a := attributeType{flags: extendedLength}
	a.setNotExtendedLength()
	if !a.nonextendedLength() {
		t.Error("Expected attribute to be extended length")
	}

	a = attributeType{flags: notExtendedLength}
	a.setNotExtendedLength()
	if !a.nonextendedLength() {
		t.Error("Expected attribute to be extended length")
	}
	// TODO: Should also test that we're not squashing other bits
}
