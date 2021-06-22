# pylint:disable-msg=E0611,I1101
"""Module bundling all functions needed to determine the date of HTML strings
or LXML trees.
"""

## This file is available from https://github.com/adbar/htmldate
## under GNU GPL v3 license

# standard
import logging
import re

from collections import Counter
from copy import deepcopy
from functools import lru_cache, partial
from lxml import etree, html

# own
from .extractors import (discard_unwanted, extract_url_date,
                         extract_partial_url_date, idiosyncrasies_search,
                         img_search, json_search, timestamp_search, try_ymd_date,
                         ADDITIONAL_EXPRESSIONS, DATE_EXPRESSIONS,
                         YEAR_PATTERN, YMD_PATTERN, COPYRIGHT_PATTERN,
                         THREE_PATTERN, THREE_CATCH,
                         THREE_LOOSE_PATTERN, THREE_LOOSE_CATCH,
                         SELECT_YMD_PATTERN, SELECT_YMD_YEAR, YMD_YEAR,
                         DATESTRINGS_PATTERN, DATESTRINGS_CATCH,
                         SLASHES_PATTERN, SLASHES_YEAR,
                         YYYYMM_PATTERN, YYYYMM_CATCH, MMYYYY_PATTERN,
                         MMYYYY_YEAR, SIMPLE_PATTERN)
from .settings import HTML_CLEANER, MAX_POSSIBLE_CANDIDATES
from .utils import load_html
from .validators import (check_extracted_reference, compare_values,
                         convert_date, date_validator, filter_ymd_candidate,
                         get_min_date, get_max_date, output_format_validator,
                         plausible_year_filter)


LOGGER = logging.getLogger(__name__)
def logstring(element):
    '''Format the element to be logged to a string.'''
    return html.tostring(element, pretty_print=False, encoding='unicode').strip()


def select_candidate(occurrences, catch, yearpat, original_date, min_date, max_date):
    """Select a candidate among the most frequent matches"""
    match = None
    # LOGGER.debug('occurrences: %s', occurrences)
    if len(occurrences) == 0 or len(occurrences) > MAX_POSSIBLE_CANDIDATES:
        return None
    if len(occurrences) == 1:
        match = catch.search(list(occurrences.keys())[0])
        if match:
            return match
    # select among most frequent
    firstselect = occurrences.most_common(10)
    LOGGER.debug('firstselect: %s', firstselect)
    # sort and find probable candidates
    if original_date is False:
        bestones = sorted(firstselect, reverse=True)[:2]
    else:
        bestones = sorted(firstselect)[:2]
    first_pattern, first_count = bestones[0][0], bestones[0][1]
    second_pattern, second_count = bestones[1][0], bestones[1][1]
    LOGGER.debug('bestones: %s', bestones)
    # same number of occurrences: always take top of the pile
    if first_count == second_count:
        match = catch.search(first_pattern)
    else:
        year1 = int(yearpat.search(first_pattern).group(1))
        year2 = int(yearpat.search(second_pattern).group(1))
        # safety net: plausibility
        if date_validator(str(year1), '%Y', earliest=min_date, latest=max_date) is False:
            if date_validator(str(year2), '%Y', earliest=min_date, latest=max_date) is True:
                # LOGGER.debug('first candidate not suitable: %s', year1)
                match = catch.search(second_pattern)
            else:
                LOGGER.debug('no suitable candidate: %s %s', year1, year2)
                return None
        # safety net: newer date but up to 50% less frequent
        if year2 != year1 and second_count/first_count > 0.5:
            match = catch.search(second_pattern)
        # not newer or hopefully not significant
        else:
            match = catch.search(first_pattern)
    return match


def search_pattern(htmlstring, pattern, catch, yearpat, original_date, min_date, max_date):
    """Chained candidate filtering and selection"""
    candidates = plausible_year_filter(htmlstring, pattern, yearpat)
    return select_candidate(candidates, catch, yearpat, original_date, min_date, max_date)


def normalize_match(match):
    '''Normalize string output by adding "0" if necessary.'''
    if len(match.group(1)) == 1:
        day = '0' + match.group(1)
    else:
        day = match.group(1)
    if len(match.group(2)) == 1:
        month = '0' + match.group(2)
    else:
        month = match.group(2)
    return day, month


def search_page(htmlstring, outputformat, original_date, min_date, max_date):
    """
    Opportunistically search the HTML text for common text patterns

    :param htmlstring:
        The HTML document in string format, potentially cleaned and stripped to
        the core (much faster)
    :type htmlstring: string
    :param outputformat:
        Provide a valid datetime format for the returned string
        (see datetime.strftime())
    :type outputformat: string
    :param original_date:
        Look for original date (e.g. publication date) instead of most recent
        one (e.g. last modified, updated time)
    :type original_date: boolean
    :return: Returns a valid date expression as a string, or None

    """

    # copyright symbol
    LOGGER.debug('looking for copyright/footer information')
    copyear = 0
    bestmatch = search_pattern(htmlstring, COPYRIGHT_PATTERN, YEAR_PATTERN, YEAR_PATTERN, original_date, min_date, max_date)
    if bestmatch is not None:
        LOGGER.debug('Copyright detected: %s', bestmatch.group(0))
        if date_validator(bestmatch.group(0), '%Y', latest=max_date) is True:
            LOGGER.debug('copyright year/footer pattern found: %s', bestmatch.group(0))
            copyear = int(bestmatch.group(0))

    # 3 components
    LOGGER.debug('3 components')
    # target URL characteristics
    bestmatch = search_pattern(htmlstring, THREE_PATTERN, THREE_CATCH, YEAR_PATTERN, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, THREE_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # more loosely structured data
    bestmatch = search_pattern(htmlstring, THREE_LOOSE_PATTERN, THREE_LOOSE_CATCH, YEAR_PATTERN, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, THREE_LOOSE_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # YYYY-MM-DD/DD-MM-YYYY
    candidates = plausible_year_filter(htmlstring, SELECT_YMD_PATTERN, SELECT_YMD_YEAR)
    # revert DD-MM-YYYY patterns before sorting
    replacement = dict()
    for item in candidates:
        match = re.match(r'([0-3]?[0-9])[/.-]([01]?[0-9])[/.-]([0-9]{4})', item)
        day, month = normalize_match(match)
        candidate = '-'.join([match.group(3), month, day])
        replacement[candidate] = candidates[item]
    candidates = Counter(replacement)
    # select
    bestmatch = select_candidate(candidates, YMD_PATTERN, YMD_YEAR, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, SELECT_YMD_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # valid dates strings
    bestmatch = search_pattern(htmlstring, DATESTRINGS_PATTERN, DATESTRINGS_CATCH, YEAR_PATTERN, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, DATESTRINGS_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # DD?/MM?/YY
    candidates = plausible_year_filter(htmlstring, SLASHES_PATTERN, SLASHES_YEAR, tocomplete=True)
    # revert DD-MM-YYYY patterns before sorting
    replacement = dict()
    for item in candidates:
        match = re.match(r'([0-3]?[0-9])[/.]([01]?[0-9])[/.]([0-9]{2})', item)
        day, month = normalize_match(match)
        if match.group(3)[0] == '9':
            year = '19' + match.group(3)
        else:
            year = '20' + match.group(3)
        candidate = '-'.join([year, month, day])
        replacement[candidate] = candidates[item]
    candidates = Counter(replacement)
    bestmatch = select_candidate(candidates, YMD_PATTERN, YMD_YEAR, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, SLASHES_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # 2 components
    LOGGER.debug('switching to two components')
    # first option
    bestmatch = search_pattern(htmlstring, YYYYMM_PATTERN, YYYYMM_CATCH, YEAR_PATTERN, original_date, min_date, max_date)
    if bestmatch is not None:
        pagedate = '-'.join([bestmatch.group(1), bestmatch.group(2), '01'])
        if date_validator(pagedate, '%Y-%m-%d', latest=max_date) is True:
            if copyear == 0 or int(bestmatch.group(1)) >= copyear:
                LOGGER.debug('date found for pattern "%s": %s', YYYYMM_PATTERN, pagedate)
                return convert_date(pagedate, '%Y-%m-%d', outputformat)

    # 2 components, second option
    candidates = plausible_year_filter(htmlstring, MMYYYY_PATTERN, MMYYYY_YEAR, original_date)
    # revert DD-MM-YYYY patterns before sorting
    replacement = dict()
    for item in candidates:
        match = re.match(r'([0-3]?[0-9])[/.-]([0-9]{4})', item)
        if len(match.group(1)) == 1:
            month = '0' + match.group(1)
        else:
            month = match.group(1)
        candidate = '-'.join([match.group(2), month, '01'])
        replacement[candidate] = candidates[item]
    candidates = Counter(replacement)
    # select
    bestmatch = select_candidate(candidates, YMD_PATTERN, YMD_YEAR, original_date, min_date, max_date)
    result = filter_ymd_candidate(bestmatch, MMYYYY_PATTERN, original_date, copyear, outputformat, min_date, max_date)
    if result is not None:
        return result

    # catchall
    if copyear != 0:
        LOGGER.debug('using copyright year as default')
        return convert_date('-'.join([str(copyear), '01', '01']), '%Y-%m-%d', outputformat)

    # 1 component, last try
    LOGGER.debug('switching to one component')
    bestmatch = search_pattern(htmlstring, SIMPLE_PATTERN, YEAR_PATTERN, YEAR_PATTERN, original_date, min_date, max_date)
    if bestmatch is not None:
        pagedate = '-'.join([bestmatch.group(0), '01', '01'])
        if date_validator(pagedate, '%Y-%m-%d', latest=max_date) is True:
            if copyear == 0 or int(bestmatch.group(0)) >= copyear:
                LOGGER.debug('date found for pattern "%s": %s', SIMPLE_PATTERN, pagedate)
                return convert_date(pagedate, '%Y-%m-%d', outputformat)

    return None


def find_date(htmlobject, extensive_search=True, original_date=False, outputformat='%Y-%m-%d', url=None, verbose=False, min_date=None, max_date=None):
    # precise patterns and idiosyncrasies
    text_result = idiosyncrasies_search(htmlstring, outputformat, min_date, max_date)
    if text_result is not None:
        return text_result

    # title
    for title_elem in cleaned_html.iterfind('.//title|.//h1'):
        attempt = try_ymd_date(title_elem.text_content(), outputformat, extensive_search, min_date, max_date)
        if attempt is not None:
            return attempt

    # last try: URL 2
    if url is not None:
        dateresult = extract_partial_url_date(url, outputformat)
        if dateresult is not None:
            return dateresult

    # try image elements
    img_result = img_search(
        tree, outputformat, min_date, max_date
        )
    if img_result is not None:
        return img_result

    # last resort
    if extensive_search is True:
        LOGGER.debug('extensive search started')
        # div and p elements?
        # TODO: check all and decide according to original_date
        reference = 0
        for textpart in [t for t in cleaned_html.xpath('.//div/text()|.//p/text()') if 0 < len(t) < 80]:
            reference = compare_reference(reference, textpart, outputformat, extensive_search, original_date, min_date, max_date)
        # return
        converted = check_extracted_reference(reference, outputformat, min_date, max_date)
        if converted is not None:
            return converted
        # search page HTML
        return search_page(htmlstring, outputformat, original_date, min_date, max_date)

    return None
