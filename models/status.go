// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"io"
	"strings"
)

type Status struct {
	State  State
	Mode   Mode
	Uptime Duration

	DownstreamActualRate ValueBandwidth
	UpstreamActualRate   ValueBandwidth

	DownstreamAttainableRate ValueBandwidth
	UpstreamAttainableRate   ValueBandwidth

	DownstreamMinimumErrorFreeThroughput ValueBandwidth
	UpstreamMinimumErrorFreeThroughput   ValueBandwidth

	DownstreamBitswapEnabled BoolValue
	UpstreamBitswapEnabled   BoolValue

	DownstreamSeamlessRateAdaption BoolValue
	UpstreamSeamlessRateAdaption   BoolValue

	DownstreamInterleavingDelay ValueMilliseconds
	UpstreamInterleavingDelay   ValueMilliseconds

	DownstreamImpulseNoiseProtection ValueSymbols
	UpstreamImpulseNoiseProtection   ValueSymbols

	DownstreamRetransmissionEnabled BoolValue
	UpstreamRetransmissionEnabled   BoolValue

	DownstreamVectoringState VectoringValue
	UpstreamVectoringState   VectoringValue

	DownstreamAttenuation ValueDecibel
	UpstreamAttenuation   ValueDecibel

	DownstreamSNRMargin ValueDecibel
	UpstreamSNRMargin   ValueDecibel

	DownstreamPower ValuePower
	UpstreamPower   ValuePower

	DownstreamFECCount IntValue
	UpstreamFECCount   IntValue

	DownstreamRTXTXCount IntValue
	UpstreamRTXTXCount   IntValue

	DownstreamRTXCCount IntValue
	UpstreamRTXCCount   IntValue

	DownstreamRTXUCCount IntValue
	UpstreamRTXUCCount   IntValue

	DownstreamCRCCount IntValue
	UpstreamCRCCount   IntValue

	DownstreamESCount IntValue
	UpstreamESCount   IntValue

	DownstreamSESCount IntValue
	UpstreamSESCount   IntValue

	FarEndInventory  Inventory
	NearEndInventory Inventory
}

func (s Status) Summary() string {
	var b strings.Builder

	fmt.Fprintf(&b, "           State:    %s\n", s.State)
	fmt.Fprintf(&b, "            Mode:    %s\n", s.Mode)
	fmt.Fprintf(&b, "          Uptime:    %s\n", s.Uptime)
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "          Remote:    %s\n", s.FarEndInventory)
	fmt.Fprintf(&b, "           Modem:    %s\n", s.NearEndInventory)
	fmt.Fprintln(&b)

	printValues(&b, "Actual rate", s.DownstreamActualRate, s.UpstreamActualRate)
	printValues(&b, "Attainable rate", s.DownstreamAttainableRate, s.UpstreamAttainableRate)
	printValues(&b, "MINEFTR", s.DownstreamMinimumErrorFreeThroughput, s.UpstreamMinimumErrorFreeThroughput)
	fmt.Fprintln(&b)

	printValues(&b, "Bitswap", s.DownstreamBitswapEnabled, s.UpstreamBitswapEnabled)
	printValues(&b, "Rate adaption", s.DownstreamSeamlessRateAdaption, s.UpstreamSeamlessRateAdaption)
	fmt.Fprintln(&b)

	printValues(&b, "Interleaving", s.DownstreamInterleavingDelay, s.UpstreamInterleavingDelay)
	printValues(&b, "INP", s.DownstreamImpulseNoiseProtection, s.UpstreamImpulseNoiseProtection)
	printValues(&b, "Retransmission", s.DownstreamRetransmissionEnabled, s.UpstreamRetransmissionEnabled)
	fmt.Fprintln(&b)

	printValues(&b, "Vectoring", s.DownstreamVectoringState, s.UpstreamVectoringState)
	fmt.Fprintln(&b)

	printValues(&b, "Attenuation", s.DownstreamAttenuation, s.UpstreamAttenuation)
	printValues(&b, "SNR margin", s.DownstreamSNRMargin, s.UpstreamSNRMargin)
	printValues(&b, "Transmit power", s.DownstreamPower, s.UpstreamPower)
	fmt.Fprintln(&b)

	printValues(&b, "FEC Count", s.DownstreamFECCount, s.UpstreamFECCount)
	printValues(&b, "RTX TX Count", s.DownstreamRTXTXCount, s.UpstreamRTXTXCount)
	printValues(&b, "RTX C Count", s.DownstreamRTXCCount, s.UpstreamRTXCCount)
	printValues(&b, "RTX UC Count", s.DownstreamRTXUCCount, s.UpstreamRTXUCCount)
	printValues(&b, "CRC Count", s.DownstreamCRCCount, s.UpstreamCRCCount)
	printValues(&b, "ES Count", s.DownstreamESCount, s.UpstreamESCount)
	printValues(&b, "SES Count", s.DownstreamSESCount, s.UpstreamSESCount)

	return b.String()
}

func printValues(w io.Writer, label string, valDown, valUp Value) {
	fmt.Fprintf(w, "%16s:    %8s %-7s  %8s %-7s\n", label, valDown.Value(), valDown.Unit(), valUp.Value(), valUp.Unit())
}
