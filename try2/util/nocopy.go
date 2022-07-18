package util

type NoCopy struct{}

func (*NoCopy) Lock() {}
