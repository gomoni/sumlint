package test

import "github.com/gomoni/sumlint/test/one_of"

type oneof struct{}

func (o oneof) good(msg *one_of.Msg) {
	switch msg.GetPayload().(type) {
	case *one_of.Msg_A:
	case *one_of.Msg_B:
	default:
	}
}

func (o oneof) nonDefault(msg *one_of.Msg) {
	switch msg.GetPayload().(type) {
	case *one_of.Msg_A:
	case *one_of.Msg_B:
	}
}

func (o oneof) noB(msg *one_of.Msg) {
	switch msg.GetPayload().(type) {
	case *one_of.Msg_A:
	default:
	}
}
