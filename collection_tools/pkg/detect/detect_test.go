// SPDX-License-Identifier: GPL-2.0-or-later

package detect

import (
	"bytes"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const phc2sysPSOutput = `USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root 31275 0.0 0.0 ? Sl 04:09 0:14 /usr/sbin/ts2phc -f /var/run/ts2phc.0.config -m
root 31379 0.1 0.0 ? S 04:09 0:46 /usr/sbin/phc2sys -f /var/run/phc2sys.0.config -r -u 0 -m -w -n 24 -s eno8703np0`

type fakeExecContext struct {
	stdout string
}

func (f *fakeExecContext) ExecCommand([]string) (stdout, stderr string, err error) {
	return f.stdout, "", nil
}

func (f *fakeExecContext) ExecCommandStdIn([]string, bytes.Buffer) (stdout, stderr string, err error) {
	return f.stdout, "", nil
}

var _ = Describe("parsePTPClockIndexFromEthtool", func() {
	It("parses the legacy PTP Hardware Clock line", func() {
		out := `Time stamping parameters for eth0:
Capabilities:
	hardware-transmit
	hardware-raw-clock
PTP Hardware Clock: 2
Hardware Transmit Timestamp Modes:
	off
	on`

		index, ok := parsePTPClockIndexFromEthtool(out)
		Expect(ok).To(BeTrue())
		Expect(index).To(Equal("2"))
	})

	It("parses the hardware timestamp provider index line", func() {
		out := `Time stamping parameters for ens3f0:
Capabilities:
	hardware-transmit
	hardware-raw-clock
Hardware timestamp provider index: 0
Hardware timestamp provider qualifier: Precise (IEEE 1588 quality)
Hardware Transmit Timestamp Modes:
	off
	on`

		index, ok := parsePTPClockIndexFromEthtool(out)
		Expect(ok).To(BeTrue())
		Expect(index).To(Equal("0"))
	})

	It("ignores PTP Hardware Clock set to none", func() {
		out := `PTP Hardware Clock: none
Hardware timestamp provider index: 1`

		index, ok := parsePTPClockIndexFromEthtool(out)
		Expect(ok).To(BeTrue())
		Expect(index).To(Equal("1"))
	})

	It("returns false when no clock index is present", func() {
		out := `Time stamping parameters for eth0:
Capabilities:
	software-transmit`

		_, ok := parsePTPClockIndexFromEthtool(out)
		Expect(ok).To(BeFalse())
	})
})

var _ = Describe("parsePhc2SysSourceInterfaces", func() {
	It("extracts the source interface of the phc2sys process", func() {
		Expect(parsePhc2SysSourceInterfaces(phc2sysPSOutput)).To(Equal([]string{"eno8703np0"}))
	})

	It("supports the -s= form of the source flag", func() {
		out := `root 31379 0.1 0.0 11332 5176 ? S 04:09 0:46 /usr/sbin/phc2sys -s=eno8703np0 -r -m`

		Expect(parsePhc2SysSourceInterfaces(out)).To(Equal([]string{"eno8703np0"}))
	})

	It("ignores phc2sys processes without a source flag", func() {
		out := `root 31379 0.1 0.0 11332 5176 ? S 04:09 0:46 /usr/sbin/phc2sys -a -r -m`

		Expect(parsePhc2SysSourceInterfaces(out)).To(BeEmpty())
	})

	It("ignores ts2phc processes", func() {
		out := `root 31275 0.0 0.0 85068 5400 ? Sl 04:09 0:14 /usr/sbin/ts2phc -f /var/run/ts2phc.0.config -s eno8703np0`

		Expect(parsePhc2SysSourceInterfaces(out)).To(BeEmpty())
	})
})

var _ = Describe("resolvePrimaryFromPhc2Sys", func() {
	It("marks the phc2sys source interface as primary when no primary exists", func() {
		ctx := &fakeExecContext{stdout: phc2sysPSOutput}
		interfaces := []DetectedInterface{
			{Name: "eno8703np0", PTPClockDevicePath: "/dev/ptp0"},
			{Name: "enp108s0f0np0", PTPClockDevicePath: "/dev/ptp1"},
		}

		result := resolvePrimaryFromPhc2Sys(ctx, interfaces)

		Expect(result[0].Primary).To(BeTrue())
		Expect(result[1].Primary).To(BeFalse())
	})

	It("leaves the interfaces untouched when a primary already exists", func() {
		ctx := &fakeExecContext{stdout: phc2sysPSOutput}
		interfaces := []DetectedInterface{
			{Name: "eno8703np0", PTPClockDevicePath: "/dev/ptp0"},
			{Name: "enp108s0f0np0", PTPClockDevicePath: "/dev/ptp1", Primary: true},
		}

		result := resolvePrimaryFromPhc2Sys(ctx, interfaces)

		Expect(result[0].Primary).To(BeFalse())
		Expect(result[1].Primary).To(BeTrue())
	})

	It("leaves the interfaces untouched when the source is not among them", func() {
		ctx := &fakeExecContext{stdout: phc2sysPSOutput}
		interfaces := []DetectedInterface{
			{Name: "ens1f0", PTPClockDevicePath: "/dev/ptp0"},
		}

		result := resolvePrimaryFromPhc2Sys(ctx, interfaces)

		Expect(result[0].Primary).To(BeFalse())
	})

	It("leaves the interfaces untouched when no phc2sys source is found", func() {
		ctx := &fakeExecContext{stdout: "USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND"}
		interfaces := []DetectedInterface{
			{Name: "ens1f0", PTPClockDevicePath: "/dev/ptp0"},
		}

		result := resolvePrimaryFromPhc2Sys(ctx, interfaces)

		Expect(result[0].Primary).To(BeFalse())
	})
})

func TestDetect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detect Suite")
}
