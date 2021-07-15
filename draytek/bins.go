// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/models"
)

var regexpBandinfo = regexp.MustCompile(`Limits=\[\s*(\d+)-\s*(\d+)\]`)

func parseBins(status models.Status, bandinfo, downstream, upstream, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode
	bins.Bins = make([]models.Bin, bins.Mode.BinCount())

	parseStatusBandinfo(&bins, bandinfo)

	parseShowbinsData(&bins, downstream)
	parseShowbinsData(&bins, upstream)

	parseStatusQLN(&bins, qln)
	parseStatusHlog(&bins, hlog)

	return bins
}

func parseStatusBandinfo(bins *models.Bins, data string) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var binType models.BinType

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "US:") {
			binType = models.BinTypeUpstream
		} else if strings.HasPrefix(line, "DS:") {
			binType = models.BinTypeDownstream
		}

		submatches := regexpBandinfo.FindStringSubmatch(line)
		if len(submatches) == 3 {
			start, _ := strconv.ParseInt(submatches[1], 10, 64)
			end, _ := strconv.ParseInt(submatches[2], 10, 64)

			for num := start; num <= end; num++ {
				bins.Bins[num].Type = binType
			}
		}
	}
}

func parseShowbinsData(bins *models.Bins, data string) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var maxSNRIndex, maxBitsIndex int
	snrData := make([]float64, bins.Mode.BinCount())

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			items := strings.Split(line, "*")
			for _, item := range items {
				readShowbinsBin(bins, snrData, &maxSNRIndex, &maxBitsIndex, item)
			}
		}
	}

	handleShowbinsSNR(bins, snrData, maxSNRIndex, maxBitsIndex)
}

func readShowbinsBin(bins *models.Bins, snrData []float64, maxSNRIndex, maxBitsIndex *int, item string) {
	data := strings.Fields(item)
	if len(data) == 4 {
		num, _ := strconv.Atoi(data[0])
		snr, _ := strconv.ParseFloat(data[1], 64)
		bits, _ := strconv.ParseInt(data[3], 10, 64)

		if bits != 0 {
			bins.Bins[num].Bits = int8(bits)
			*maxBitsIndex = num
		}
		if snr != 0 {
			snrData[num] = snr
			*maxSNRIndex = num
		}
	}
}

func handleShowbinsSNR(bins *models.Bins, snrData []float64, maxSNRIndex, maxBitsIndex int) {
	if maxSNRIndex == 0 {
		return
	}

	if maxBitsIndex > 512 {

		// There is a bug in the bin data for at least some VDSL firmwares: The SNR data of the entire
		// frequency range is stored in the bins 0-511, with all others being zero.

		maxFactor := bins.Mode.BinCount() / maxSNRIndex
		var factor int
		for factor = 1; factor < maxFactor; factor *= 2 {
			// after applying factor, maxSNRIndex should be at most 10% lower than maxBitsIndex, because:
			// - maxSNRIndex > maxBitsIndex is common when SNR is too low to allocate bits
			// - maxSNRIndex < maxBitsIndex unlikely, as SNR value needed to allocate bins
			if float64(maxSNRIndex*factor)/float64(maxBitsIndex) > 0.9 {
				break
			}
		}

		for i := 0; i <= maxSNRIndex; i++ {
			val := snrData[i]
			if val != 0 {
				numBase := i * factor
				for num := numBase; num < numBase+factor; num++ {
					bins.Bins[num].SNR = val
				}
			}
		}

	} else {

		for num, val := range snrData {
			if val != 0 {
				bins.Bins[num].SNR = val
			}
		}

	}
}

func parseStatusQLN(bins *models.Bins, qln string) {
	parseStatusBins(qln, func(num int, val float64, ok bool) {
		if ok && val != -150 {
			bins.Bins[num].QLN = val
		}
	})
}

func parseStatusHlog(bins *models.Bins, hlog string) {
	for num := range bins.Bins {
		bins.Bins[num].Hlog = -96.3
	}

	parseStatusBins(hlog, func(num int, val float64, ok bool) {
		if ok {
			bins.Bins[num].Hlog = val
		}
	})
}

func parseStatusBins(data string, handler func(int, float64, bool)) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var groupSize, groupSizeDS, groupSizeUS int

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "GroupSize") {
			if indexColon := strings.IndexRune(line, ':'); indexColon != -1 {
				lineSplit := strings.Fields(line[indexColon+1:])
				if len(lineSplit) >= 2 {
					groupSizeDS, _ = strconv.Atoi(lineSplit[0])
					groupSizeUS, _ = strconv.Atoi(lineSplit[1])
				}
			}
			continue
		}

		if strings.HasPrefix(line, "US:") {
			groupSize = groupSizeDS
		} else if strings.HasPrefix(line, "DS:") {
			groupSize = groupSizeUS
		}

		if strings.HasPrefix(line, "bin=") && groupSize != 0 {
			readStatusBin(line[4:], groupSize, handler)
		}
	}
}

func readStatusBin(line string, groupSize int, handler func(int, float64, bool)) {
	lineSplit := strings.SplitN(line, ":", 2)
	if len(lineSplit) == 2 {
		numBaseStr := strings.TrimSpace(lineSplit[0])
		numBase, _ := strconv.Atoi(numBaseStr)

		dataSplit := strings.Split(lineSplit[1], ",")
		for i := 0; i < len(dataSplit)-1; i++ {
			valStr := strings.TrimSpace(dataSplit[i])
			val, err := strconv.ParseFloat(valStr, 64)
			ok := err == nil
			numGroup := (numBase + i) * groupSize

			for j := 0; j < groupSize; j++ {
				num := numGroup + j
				handler(num, val, ok)
			}
		}
	}
}
