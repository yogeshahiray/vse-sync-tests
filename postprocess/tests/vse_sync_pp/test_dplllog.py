### SPDX-License-Identifier: GPL-2.0-or-later

"""Test cases for vse_sync_pp.dplllog"""

from unittest import TestCase

from vse_sync_pp.dplllog import parse_dpll_log_line


class TestParseDPLLLogLine(TestCase):
    """Test cases for vse_sync_pp.dplllog.parse_dpll_log_line"""

    daemon_line = (
        'dpll[1792510667]:[ts2phc.0.config] eno8703np0 '
        'frequency_status 3 offset -2 phase_status 3 pps_status 1 s2\n'
    )

    def test_parses_daemon_log_line(self):
        self.assertEqual(
            parse_dpll_log_line(self.daemon_line),
            ['1792510667', 3, 3, -2],
        )

    def test_parses_line_without_configuration_name(self):
        self.assertEqual(
            parse_dpll_log_line(
                'dpll[1792510667]: eno8703np0 frequency_status 3 offset 1 '
                'phase_status 3 pps_status 1 s2\n'),
            ['1792510667', 3, 3, 1],
        )

    def test_negative_and_positive_offsets(self):
        self.assertEqual(
            parse_dpll_log_line(
                'dpll[42]:[ts2phc.0.config] ens1f0 frequency_status -1 '
                'offset 17 phase_status -1 pps_status 0 s0\n'),
            ['42', -1, -1, 17],
        )

    def test_interface_filter(self):
        self.assertIsNone(
            parse_dpll_log_line(self.daemon_line, interface='ens1f0'))
        self.assertEqual(
            parse_dpll_log_line(self.daemon_line, interface='eno8703np0'),
            ['1792510667', 3, 3, -2],
        )

    def test_discards_non_dpll_lines(self):
        for line in (
            'gnss[1792510667]:[ts2phc.0.config] eno8703np0 gnss_status 3 offset 4 s2\n',
            'ts2phc[41790.480]: [ts2phc.0.config:7] nmea sentence: GNRMC\n',
            'I1020 15:37:47.406441 10632 dpll.go:433] setting phase offset to -2 ns\n',
            'phc2sys[1792523048.813]: [ptp4l.0.config:6] phc offset summary\n',
            '',
        ):
            self.assertIsNone(parse_dpll_log_line(line))
