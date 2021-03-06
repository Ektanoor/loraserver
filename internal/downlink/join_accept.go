package downlink

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/loraserver/internal/common"
	"github.com/brocaar/loraserver/internal/storage"
)

func getJoinAcceptTXInfo(ctx *JoinContext) error {
	if len(ctx.DeviceSession.LastRXInfoSet) == 0 {
		return errors.New("empty LastRXInfoSet")
	}

	rxInfo := ctx.DeviceSession.LastRXInfoSet[0]

	ctx.TXInfo = gw.TXInfo{
		MAC:      rxInfo.MAC,
		CodeRate: rxInfo.CodeRate,
		Power:    common.Band.DefaultTXPower,
	}

	var timestamp uint32

	if ctx.DeviceSession.RXWindow == storage.RX1 {
		timestamp = rxInfo.Timestamp + uint32(common.Band.JoinAcceptDelay1/time.Microsecond)

		// get uplink dr
		uplinkDR, err := common.Band.GetDataRate(rxInfo.DataRate)
		if err != nil {
			return errors.Wrap(err, "get data-rate error")
		}

		// get RX1 DR
		rx1DR, err := common.Band.GetRX1DataRate(uplinkDR, 0)
		if err != nil {
			return errors.Wrap(err, "get rx1 data-rate error")
		}
		ctx.TXInfo.DataRate = common.Band.DataRates[rx1DR]

		// get RX1 frequency
		ctx.TXInfo.Frequency, err = common.Band.GetRX1Frequency(rxInfo.Frequency)
		if err != nil {
			return errors.Wrap(err, "get rx1 frequency error")
		}
	} else if ctx.DeviceSession.RXWindow == storage.RX2 {
		timestamp = rxInfo.Timestamp + uint32(common.Band.JoinAcceptDelay2/time.Microsecond)
		ctx.TXInfo.DataRate = common.Band.DataRates[common.Band.RX2DataRate]
		ctx.TXInfo.Frequency = common.Band.RX2Frequency
	} else {
		return fmt.Errorf("unknown RXWindow defined %d", ctx.DeviceSession.RXWindow)
	}

	ctx.TXInfo.Timestamp = &timestamp

	return nil
}

func logJoinAcceptFrame(ctx *JoinContext) error {
	logDownlink(common.DB, ctx.DeviceSession.DevEUI, ctx.PHYPayload, ctx.TXInfo)
	return nil
}

func sendJoinAcceptResponse(ctx *JoinContext) error {
	err := common.Gateway.SendTXPacket(gw.TXPacket{
		TXInfo:     ctx.TXInfo,
		PHYPayload: ctx.PHYPayload,
	})
	if err != nil {
		return errors.Wrap(err, "send tx-packet error")
	}

	return nil
}
