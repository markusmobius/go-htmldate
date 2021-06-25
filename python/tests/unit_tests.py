# pylint:disable-msg=W1401
"""
Unit tests for the htmldate library.
"""


import datetime
import logging
import os
import re
import sys

import pytest

from collections import Counter
from unittest.mock import patch

try:
    import dateparser
    EXT_PARSER = True
    PARSER = dateparser.DateDataParser(languages=['de', 'en'], settings={'PREFER_DAY_OF_MONTH': 'first', 'PREFER_DATES_FROM': 'past', 'DATE_ORDER': 'DMY'}) # allow_redetect_language=False,
except ImportError:
    EXT_PARSER = False

from lxml import html

try:
    import cchardet as chardet
except ImportError:
    import chardet

from htmldate.cli import examine, parse_args
from htmldate.core import compare_reference, find_date, search_page, search_pattern, select_candidate, try_ymd_date
from htmldate.extractors import custom_parse, external_date_parser, extract_partial_url_date, regex_parse
from htmldate.settings import MIN_DATE, MIN_DATE, LATEST_POSSIBLE
from htmldate.utils import fetch_url, load_html
from htmldate.validators import convert_date, date_validator, get_max_date, get_min_date, output_format_validator


TEST_DIR = os.path.abspath(os.path.dirname(__file__))
OUTPUTFORMAT = '%Y-%m-%d'

logging.basicConfig(stream=sys.stdout, level=logging.DEBUG)
# '': '', \



# def new_pages():
#    '''New pages, to be sorted'''
    # assert find_date(load_mock_page('...')) == 'YYYY-MM-DD'
#    pass

def test_search_pattern(original_date=False, min_date=MIN_DATE, max_date=LATEST_POSSIBLE):
    '''test pattern search in strings'''
    #
    pattern = re.compile(r'\D([0-9]{4}[/.-][0-9]{2})\D')
    catch = re.compile(r'([0-9]{4})[/.-]([0-9]{2})')
    yearpat = re.compile(r'^([12][0-9]{3})')
    assert search_pattern('It happened on the 202.E.19, the day when it all began.', pattern, catch, yearpat, original_date, min_date, max_date) is None
    assert search_pattern('The date is 2002.02.15.', pattern, catch, yearpat, original_date, min_date, max_date) is not None
    assert search_pattern('http://www.url.net/index.html', pattern, catch, yearpat, original_date, min_date, max_date) is None
    assert search_pattern('http://www.url.net/2016/01/index.html', pattern, catch, yearpat, original_date, min_date, max_date) is not None
    #
    pattern = re.compile(r'\D([0-9]{2}[/.-][0-9]{4})\D')
    catch = re.compile(r'([0-9]{2})[/.-]([0-9]{4})')
    yearpat = re.compile(r'([12][0-9]{3})$')
    assert search_pattern('It happened on the 202.E.19, the day when it all began.', pattern, catch, yearpat, original_date, min_date, max_date) is None
    assert search_pattern('It happened on the 15.02.2002, the day when it all began.', pattern, catch, yearpat, original_date, min_date, max_date) is not None
    #
    pattern = re.compile(r'\D(2[01][0-9]{2})\D')
    catch = re.compile(r'(2[01][0-9]{2})')
    yearpat = re.compile(r'^(2[01][0-9]{2})')
    assert search_pattern('It happened in the film 300.', pattern, catch, yearpat, original_date, min_date, max_date) is None
    assert search_pattern('It happened in 2002.', pattern, catch, yearpat, original_date, min_date, max_date) is not None


def test_dependencies():
    '''Test README example for consistency'''
    if EXT_PARSER is True:
        assert try_ymd_date('Fri | September 1 | 2017', OUTPUTFORMAT, True, MIN_DATE, LATEST_POSSIBLE) == '2017-09-01'
