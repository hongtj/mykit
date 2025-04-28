package transfer

import (
	. "mykit/core/dsp"
	. "mykit/core/types"
)

func RegLogSenderGrpc(address string, must ...bool) (res LogSenderMaker) {
	res = func() (mode int, sender LogSender, err error) {
		mode = LogSenderGrpc
		sender, err = NewGrpcLogSender(address)

		if err != nil {
			HandleInitErr("grpc log sender init", err, must...)
		}

		return
	}

	RegLogSender(res)

	return
}
