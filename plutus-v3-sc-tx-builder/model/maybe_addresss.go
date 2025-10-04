package model

import "github.com/Salvionied/apollo/serialization/Address"

func (Nothing) MaybeAddress()     {}
func (WithAddress) MaybeAddress() {}

type MaybeAddress interface {
	MaybeAddress()
}

type WithAddress struct {
	_       struct{}        `plutusType:"DefList" plutusConstr:"0"`
	Address Address.Address `plutusType:"Address"`
}
