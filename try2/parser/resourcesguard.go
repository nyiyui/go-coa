package parser

import "gitlab.com/coalang/go-coa/try2/util"

type ResourcesGuard interface {
	Allowed(util.Resource) bool
}

type PureRG struct{}

var _ ResourcesGuard = PureRG{}

func (p PureRG) Allowed(r util.Resource) bool { return false }

type MapRG struct{ m map[util.Resource]bool }

var _ ResourcesGuard = new(MapRG)

func (m *MapRG) Allowed(resource util.Resource) bool { return m.m[resource] }
