# pylint:disable-msg=E0611,I1101
"""
Filters for date parsing and date validators.
"""

## This file is available from https://github.com/adbar/htmldate
## under GNU GPL v3 license

# standard
import datetime
import logging
import time

from collections import Counter
from functools import lru_cache

from .settings import MIN_DATE, MIN_YEAR, LATEST_POSSIBLE, MAX_YEAR


LOGGER = logging.getLogger(__name__)
LOGGER.debug('date settings: %s %s %s', MIN_YEAR, LATEST_POSSIBLE, MAX_YEAR)


def output_format_validator(outputformat):
    """Validate the output format in the settings"""
    # test in abstracto
    if not isinstance(outputformat, str) or not '%' in outputformat:
        logging.error('malformed output format: %s', outputformat)
        return False
    # test with date object
    dateobject = datetime.datetime(2017, 9, 1, 0, 0)
    try:
        dateobject.strftime(outputformat)
    except (NameError, TypeError, ValueError) as err:
        logging.error('wrong output format or format type: %s %s', outputformat, err)
        return False
    return True


@lru_cache(maxsize=32)
def plausible_year_filter(htmlstring, pattern, yearpat, tocomplete=False):
    """Filter the date patterns to find plausible years only"""
    # slow!
    allmatches = pattern.findall(htmlstring)
    occurrences = Counter(allmatches)
    toremove = set()
    # LOGGER.debug('occurrences: %s', occurrences)
    for item in occurrences.keys():
        # scrap implausible dates
        try:
            if tocomplete is False:
                potential_year = int(yearpat.search(item).group(1))
            else:
                lastdigits = yearpat.search(item).group(1)
                if lastdigits[0] == '9':
                    potential_year = int('19' + lastdigits)
                else:
                    potential_year = int('20' + lastdigits)
        except AttributeError:
            LOGGER.debug('not a year pattern: %s', item)
            toremove.add(item)
        else:
            if potential_year < MIN_YEAR or potential_year > MAX_YEAR:
                LOGGER.debug('no potential year: %s', item)
                toremove.add(item)
            # occurrences.remove(item)
            # continue
    # preventing dictionary changed size during iteration error
    for item in toremove:
        del occurrences[item]
    return occurrences


@lru_cache(maxsize=32)
def filter_ymd_candidate(bestmatch, pattern, original_date, copyear, outputformat, min_date, max_date):
    """Filter free text candidates in the YMD format"""
    if bestmatch is not None:
        pagedate = '-'.join([bestmatch.group(1), bestmatch.group(2), bestmatch.group(3)])
        if date_validator(pagedate, '%Y-%m-%d', earliest=min_date, latest=max_date) is True:
            if copyear == 0 or int(bestmatch.group(1)) >= copyear:
                LOGGER.debug('date found for pattern "%s": %s', pattern, pagedate)
                return convert_date(pagedate, '%Y-%m-%d', outputformat)
            ## TODO: test and improve
            #if original_date is True:
            #    if copyear == 0 or int(bestmatch.group(1)) <= copyear:
            #        LOGGER.debug('date found for pattern "%s": %s', pattern, pagedate)
            #        return convert_date(pagedate, '%Y-%m-%d', outputformat)
            #else:
            #    if copyear == 0 or int(bestmatch.group(1)) >= copyear:
            #        LOGGER.debug('date found for pattern "%s": %s', pattern, pagedate)
            #        return convert_date(pagedate, '%Y-%m-%d', outputformat)
    return None


def convert_date(datestring, inputformat, outputformat):
    """Parse date and return string in desired format"""
    # speed-up (%Y-%m-%d)
    if inputformat == outputformat:
        return str(datestring)
    # date object (speedup)
    if isinstance(datestring, datetime.date):
        return datestring.strftime(outputformat)
    # normal
    dateobject = datetime.datetime.strptime(datestring, inputformat)
    return dateobject.strftime(outputformat)