### SPDX-License-Identifier: GPL-2.0-or-later

"""Convert linuxptp-daemon DPLL monitoring log messages to canonical data."""

from argparse import ArgumentParser
import re
import sys

from .common import (
    open_input,
    print_loj,
)


# linuxptp-daemon DPLL monitoring log message, for example
# dpll[1792510667]:[ts2phc.0.config] eno8703np0 frequency_status 3 offset -2 phase_status 3 pps_status 1 s2
DPLL_LOG_REGEXP = re.compile(
    r'^dpll\[(\d+)\]:'  # unix timestamp in seconds
    r'(?:\[[^\]]*\])?\s'  # optional configuration file name
    r'(\S+)\s'  # interface name
    r'frequency_status\s+(-?\d+)\s+'
    r'offset\s+(-?\d+)\s+'
    r'phase_status\s+(-?\d+)\s+'
    r'pps_status\s+(-?\d+)\s+'
    r's(-?\d+)'
)


def parse_dpll_log_line(line, interface=None):
    """Return canonical 'dpll/time-error' data parsed from daemon log `line`.

    Return None if `line` is not a linuxptp-daemon DPLL monitoring log
    message, or does not relate to `interface` when it is given.

    The returned array ordering matches the 'dpll/time-error' parser:
    (timestamp, eecstate, state, terror). The state is taken from the DPLL
    phase lock status which uses the DPLL netlink lock status enumeration,
    the same values the analyzers filter on (2 locked, 3 locked holdover
    acquired). The eecstate is taken from the frequency status.
    """
    matched = DPLL_LOG_REGEXP.match(line)
    if matched is None:
        return None

    if interface is not None and matched.group(2) != interface:
        return None

    (timestamp, _interface, frequency_status, offset,
     phase_status, _pps_status, _servo_state) = matched.groups()

    return [timestamp, int(frequency_status), int(phase_status), int(offset)]


def main():
    """Convert linuxptp-daemon DPLL monitoring log messages to canonical data.

    Parse DPLL monitoring messages from the linuxptp-daemon logs in input
    and for each print the canonical 'dpll/time-error' data as JSON.
    """
    aparser = ArgumentParser(description=main.__doc__)
    aparser.add_argument(
        'input',
        help="input file, or '-' to read from stdin",
    )
    aparser.add_argument(
        '--interface',
        help="only convert messages for this interface",
    )
    args = aparser.parse_args()

    with open_input(args.input) as fid:
        for line in fid:
            data = parse_dpll_log_line(line, args.interface)

            # Python exits with error code 1 on EPIPE
            if data is not None and not print_loj(data):
                sys.exit(1)


if __name__ == '__main__':
    main()
