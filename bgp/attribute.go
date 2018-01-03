package bgp

func (a *attributeType) optional() bool {
	return a.flags&optional == optional
}

func (a *attributeType) setOptional() {
	a.flags = a.flags | optional
}

func (a *attributeType) wellKnown() bool {
	return a.flags&optional == wellKnown
}

func (a *attributeType) setWellKnown() {
	a.flags = a.flags &^ optional
}

func (a *attributeType) transitive() bool {
	return a.flags&transitive == transitive
}

func (a *attributeType) setTransitive() {
	a.flags = a.flags | transitive
}

func (a *attributeType) nonTransitive() bool {
	return a.flags&transitive == nonTransitive
}

func (a *attributeType) setNonTransitive() {
	a.flags = a.flags &^ transitive
}

func (a *attributeType) partial() bool {
	return a.flags&partial == partial
}

func (a *attributeType) setPartial() {
	a.flags = a.flags | partial
}

func (a *attributeType) complete() bool {
	return a.flags&partial == complete
}

func (a *attributeType) setComplete() {
	a.flags = a.flags &^ partial
}

func (a *attributeType) extendedLength() bool {
	return a.flags&extendedLength == extendedLength
}

func (a *attributeType) setExtendedLength() {
	a.flags = a.flags | extendedLength
}

func (a *attributeType) notExtendedLength() bool {
	return a.flags&extendedLength == notExtendedLength
}

func (a *attributeType) setNotExtendedLength() {
	a.flags = a.flags &^ extendedLength
}
