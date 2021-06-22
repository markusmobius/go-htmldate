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


def img_search(tree, outputformat, min_date, max_date):
    '''Skim through image elements'''
    element = tree.find('.//meta[@property="og:image"]')
    if element is not None and 'content' in element.attrib:
        result = extract_url_date(element.get('content'), outputformat)
        if result is not None and date_validator(result, outputformat, earliest=min_date, latest=max_date) is True:
            return result
    return None

