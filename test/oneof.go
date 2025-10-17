package test

import "github.com/gomoni/sumlint/test/one_of"

type oneofTest struct{}

func (oneofTest) good(msg *one_of.Msg) {
	switch msg.GetPayload().(type) {
	case *one_of.Msg_A:
	case *one_of.Msg_B:
	default:
	}
}

func (oneofTest) noDefault(msg *one_of.Msg) {
	switch msg.GetPayload().(type) {
	case *one_of.Msg_A:
	case *one_of.Msg_B:
	}
}

func (oneofTest) noB(msg *one_of.Msg) {
	payload := msg.GetPayload()
	switch payload.(type) {
	case *one_of.Msg_A:
	default:
	}
}
