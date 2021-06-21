# pylint:disable-msg=E0611,I1101
"""
Custom parsers and XPath expressions for date extraction
"""
## This file is available from https://github.com/adbar/htmldate
## under GNU GPL v3 license

# standard
import datetime
import logging
import re

from functools import lru_cache

# conditional imports with fallbacks for compatibility
# coverage for date parsing
#try:
import dateparser  # third-party, slow
EXTERNAL_PARSER = dateparser.DateDataParser(settings={
    'PREFER_DAY_OF_MONTH': 'first', 'PREFER_DATES_FROM': 'past',
    'DATE_ORDER': 'DMY',
})

# potential regex speedup
#try:
import regex
#except ImportError:
#    regex = re

# allow_redetect_language=False, languages=['de', 'en'],
#EXTERNAL_PARSER_CONFIG = {
#    'PREFER_DAY_OF_MONTH': 'first', 'PREFER_DATES_FROM': 'past',
#    'DATE_ORDER': 'DMY'
#}
#except ImportError:
#    # try dateutil parser
#    from dateutil.parser import parse as FULL_PARSE
#    EXTERNAL_PARSER = None
#    DEFAULT_PARSER_PARAMS = {'dayfirst': True, 'fuzzy': False}
#else:
#FULL_PARSE = DEFAULT_PARSER_PARAMS = None
# iso date parsing speedup
try:
    from ciso8601 import parse_datetime, parse_datetime_as_naive
except ImportError:
    #if not FULL_PARSE:
    from dateutil.parser import parse as FULL_PARSE
    parse_datetime = parse_datetime_as_naive = FULL_PARSE  # shortcut

# own
from .validators import convert_date, date_validator

LOGGER = logging.getLogger(__name__)

def discard_unwanted(tree):
    '''Delete unwanted sections of an HTML document and return them as a list'''
    my_discarded = []
    for expr in DISCARD_EXPRESSIONS:
        for subtree in tree.xpath(expr):
            my_discarded.append(subtree)
            subtree.getparent().remove(subtree)
    return tree, my_discarded


def extract_partial_url_date(testurl, outputformat):
    """Extract an approximate date out of an URL string in Y-M format"""
    match = PARTIAL_URL.search(testurl)
    if match:
        dateresult = match.group(0) + '/01'
        LOGGER.debug('found partial date in URL: %s', dateresult)
        try:
            dateobject = datetime.datetime(int(match.group(1)),
                                           int(match.group(2)),
                                           1)
            if date_validator(dateobject, outputformat) is True:
                return dateobject.strftime(outputformat)
        except ValueError as err:
            LOGGER.debug('conversion error: %s %s', dateresult, err)
    return None


def regex_parse(string):
    """Full-text parse using a series of regular expressions"""
    dateobject = None
    dateobject = regex_parse_de(string)
    if dateobject is None:
        dateobject = regex_parse_multilingual(string)
    return dateobject


def regex_parse_de(string):
    """Try full-text parse for German date elements"""
    # text match
    match = GERMAN_TEXTSEARCH.search(string)
    if not match:
        return None
    # second element
    secondelem = TEXT_MONTHS[match.group(2)]
    # process and return
    try:
        dateobject = datetime.date(int(match.group(3)),
                                   int(secondelem),
                                   int(match.group(1)))
    except ValueError:
        return None
    LOGGER.debug('German text parse: %s', dateobject)
    return dateobject


def regex_parse_multilingual(string):
    """Try full-text parse for English date elements"""
    # https://github.com/vi3k6i5/flashtext ?
    # numbers
    match = ENGLISH_DATE.search(string)
    if match:
        day, month, year = match.group(2), match.group(1), match.group(3)
    else:
        # general search
        if not GENERAL_TEXTSEARCH.search(string):
            return None
        # American English
        match = MDY_PATTERN.search(string)
        if match:
            day, month, year = match.group(2), TEXT_MONTHS[match.group(1)], \
                               match.group(4)
        # multilingual day-month-year pattern
        else:
            match = DMY_PATTERN.search(string)
            if match:
                day, month, year = match.group(1), TEXT_MONTHS[match.group(4)], \
                                   match.group(5)
            else:
                return None
    # process and return
    if len(year) == 2:
        year = '20' + year
    try:
        dateobject = datetime.date(int(year), int(month), int(day))
    except ValueError:
        return None
    LOGGER.debug('English text parse: %s', dateobject)
    return dateobject


def external_date_parser(string, outputformat):
    """Use dateutil parser or dateparser module according to system settings"""
    LOGGER.debug('send to external parser: %s', string)
    try:
        # dateparser installed or not
        #if EXTERNAL_PARSER is not None:
        target = EXTERNAL_PARSER.get_date_data(string)['date_obj']
        #else:
        #    target = FULL_PARSE(string, **DEFAULT_PARSER_PARAMS)
    # 2 types of errors possible
    except (OverflowError, ValueError):
        target = None
    # issue with data type
    if target is not None:
        return datetime.date.strftime(target, outputformat)
    return None


def img_search(tree, outputformat, min_date, max_date):
    '''Skim through image elements'''
    element = tree.find('.//meta[@property="og:image"]')
    if element is not None and 'content' in element.attrib:
        result = extract_url_date(element.get('content'), outputformat)
        if result is not None and date_validator(result, outputformat, earliest=min_date, latest=max_date) is True:
            return result
    return None


def json_search(tree, outputformat, original_date, min_date, max_date):
    '''Look for JSON time patterns in JSON sections of the tree'''
    # determine pattern
    if original_date is True:
        json_pattern = JSON_PATTERN_PUBLISHED
    else:
        json_pattern = JSON_PATTERN_MODIFIED
    # look throughout the HTML tree
    for elem in tree.xpath('.//script[@type="application/ld+json"]|//script[@type="application/settings+json"]'):
        if not elem.text or not '"date' in elem.text:
            continue
        json_match = json_pattern.search(elem.text)
        if json_match and date_validator(json_match.group(1), '%Y-%m-%d', earliest=min_date, latest=max_date):
            LOGGER.debug('JSON time found: %s', json_match.group(0))
            return convert_date(json_match.group(1), '%Y-%m-%d', outputformat)
    return None


def timestamp_search(htmlstring, outputformat, min_date, max_date):
    '''Look for timestamps throughout the web page'''
    tstamp_match = TIMESTAMP_PATTERN.search(htmlstring)
    if tstamp_match and date_validator(tstamp_match.group(1), '%Y-%m-%d', earliest=min_date, latest=max_date):
        LOGGER.debug('time regex found: %s', tstamp_match.group(0))
        return convert_date(tstamp_match.group(1), '%Y-%m-%d', outputformat)
    return None


def extract_idiosyncrasy(idiosyncrasy, htmlstring, outputformat, min_date, max_date):
    '''Look for a precise pattern throughout the web page'''
    candidate = None
    match = idiosyncrasy.search(htmlstring)
    groups = [0, 1, 2, 3] if match and match.group(3) else [] #because len(None) has no len
    try:
        groups = [0, 4, 5, 6] if match and match.group(6) else groups #because len(None) has no len
    except IndexError:
        pass
    if match and groups: #because len(None) has no len
        if match.group(1) is not None and len(match.group(1)) == 4:
            candidate = datetime.date(int(match.group(groups[1])),
                                      int(match.group(groups[2])),
                                      int(match.group(groups[3])))
        elif len(match.group(groups[3])) in (2, 4):
            # switch to MM/DD/YY
            if int(match.group(groups[2])) > 12:
                tmp1, tmp2 = groups[1], groups[2]
                groups[1], groups[2] = tmp2, tmp1
            # DD/MM/YY
            try:
                if len(match.group(groups[3])) == 2:
                    candidate = datetime.date(int('20' + match.group(groups[3])),
                                              int(match.group(groups[2])),
                                              int(match.group(groups[1])))
                else:
                    candidate = datetime.date(int(match.group(groups[3])),
                                              int(match.group(groups[2])),
                                              int(match.group(groups[1])))
            except ValueError:
                LOGGER.debug('value error in idiosyncrasies: %s', match.group(0))
    if candidate is not None:
        if date_validator(candidate, '%Y-%m-%d', earliest=min_date, latest=max_date) is True:
            LOGGER.debug('idiosyncratic pattern found: %s', match.group(0))
            return convert_date(candidate, '%Y-%m-%d', outputformat)
    return None


def idiosyncrasies_search(htmlstring, outputformat, min_date, max_date):
    '''Look for author-written dates throughout the web page'''
    result = None
    # DE
    result = extract_idiosyncrasy(DE_PATTERNS, htmlstring, outputformat, min_date, max_date)
    # EN
    if result is None:
        result = extract_idiosyncrasy(EN_PATTERNS, htmlstring, outputformat, min_date, max_date)
    # TR
    if result is None:
        result = extract_idiosyncrasy(TR_PATTERNS, htmlstring, outputformat, min_date, max_date)
    return result
