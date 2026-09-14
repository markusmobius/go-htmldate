"""Regenerate HtmlDate expectations from pinned Python, without a Go oracle."""

import argparse
from collections import Counter
from datetime import datetime
from functools import partial
import hashlib
import importlib.metadata
import json
import logging
import os
from pathlib import Path
import platform
import subprocess
import sys
import time
from types import SimpleNamespace
import warnings


ROOT = Path(__file__).resolve().parents[2]
UPSTREAM = "b8952828329abaeeb3be21387b526f2be614ce67"
SOURCES = {
    "__init__.py": "a19a9fb3210b69fb3786b22a355a804e14abdc1e21f9403051f9e2e0a499869c",
    "cli.py": "88f577f0ad1b585955d4f9391b5c0b635e006ab00c9847cfbc443a3754aa6ccb",
    "core.py": "a77ff0dcd3533da0a64c1fde9cf100e5665b93836acb0e3a1e8204a3f80e9d03",
    "extractors.py": "389685498353db3a9b5235b394e6221171ac1075b89467fb1798e510a534111e",
    "meta.py": "55760f7a0c702e84debbd30bc15c22fd05d11ce759f3d0fea4515a12ab96155d",
    "settings.py": "e803e13bb05d958a208c06ad24e06958edf634571f637bd301d2871bb8efb3ef",
    "utils.py": "4a8a35b69ea26b57fff312108f648c0551b708b6550d53a9482780f90f58e2dd",
    "validators.py": "37c8a952e5b240c24ee75765e76a6faabb692209e11528b0da22411531264900",
}
FROZEN = datetime(2026, 9, 13, 12)


def configure_reference(context):
    import dateutil.parser._parser as parser_module
    import dateutil.tz.tz as timezone_module
    import htmldate.extractors as extractors

    names, offsets = (("UTC", "UTC"), (0, 0)) if context == "utc" else (
        ("Eastern Standard Time", "Eastern Daylight Time"), (-18000, -14400)
    )
    snapshot = SimpleNamespace(tzname=names, timezone=-offsets[0], altzone=-offsets[1],
                               daylight=int(offsets[0] != offsets[1]), localtime=time.localtime)
    parser_module.time = timezone_module.time = snapshot
    parser_module.DEFAULTPARSER.info._year = 2026
    parser_module.DEFAULTPARSER.info._century = 2000
    extractors.EXTERNAL_PARSER._settings.RELATIVE_BASE = FROZEN
    extractors.dateutil_parse = partial(parser_module.parse, default=FROZEN.replace(hour=0))
    return {"zone": "America/New_York" if context == "windows_eastern" else "UTC",
            "names": names, "offsets": offsets, "parser_year": 2026}


def local_cases(context, environment):
    import dateutil.parser._parser as parser_module
    import htmldate.extractors as extractors
    from htmldate.validators import is_valid_date

    expected_timestamp = 1730611800 if context == "windows_eastern" else 1730597400
    if datetime(2024, 11, 3, 1, 30).timestamp() != expected_timestamp:
        raise RuntimeError("unexpected C datetime local timestamp environment")
    cases = []
    for fold in (0, 1):
        default = datetime(2024, 11, 3, fold=fold)
        extractors.dateutil_parse = partial(parser_module.parse, default=default)
        offset = ("-05:00" if fold else "-04:00") if context == "windows_eastern" else "+00:00"
        for suffix in ("EST", "EDT", "XYZ", "-0400", ""):
            text = ("2024 November 3 01:30 " + suffix).strip()
            for hour in (5, 6):
                earliest = f"2024-11-03T{hour:02}:15:00+00:00"
                latest = f"2024-11-03T{hour:02}:45:00+00:00"
                is_valid_date.cache_clear()
                result = extractors.custom_parse(text, "%Y-%m-%d", datetime.fromisoformat(earliest), datetime.fromisoformat(latest))
                cases.append({"kind": "fast", "input": text, "file": "", "sha256": "",
                              "audit_id": f"{context}-{fold}-{suffix}-{hour}",
                              "environment": environment,
                              "current_time": f"2024-11-03T01:30:00{offset}",
                              "default_fold": bool(fold),
                              "options": {"min": earliest, "max": latest, "fast": True, "original": False},
                              "expected": {"date": result or "", "error": ""}})
    return cases


def shortcut_cases():
    inputs = ["2017123", "201701", "2017012", "201709011234", "2020-W53", "2020-W53-7",
              "2020W537", "2021-W01-1", "2021-W53-1", "2020-W53-7T24:00",
              "\uff12\uff10\uff11\uff17\uff10\uff19\uff10\uff11", "\u0662\u0660\u0661\u0667\u0660\u0669\u0660\u0661",
              "2017\u0660\u0669\u0660\u0661", "\u00b2\u2070\u00b9\u20772017", "2020\u2003-01-02",
              "2017-09-01" + "\u00b2" * 11, "\x1c2020-01-02\x1f", "1/1/20"]
    for kind in ("fast", "try"):
        for index, text in enumerate(inputs):
            yield {"kind": kind, "input": text, "file": "", "sha256": "", "audit_id": f"shortcut-{kind}-{index}",
                   "options": {"min": "1995-01-01T00:00:00Z", "max": "2026-09-13T23:59:59.999999Z", "fast": True, "original": False}}
    for index, (text, earliest, latest) in enumerate([
        ("2020-W53-7T23:30:00-02:00", "2021-01-04T00:00:00Z", "2021-01-04T02:00:00Z"),
        ("2020-W53-7T00:30:00+02:00", "2021-01-03T00:00:00Z", "2021-01-03T23:59:59Z"),
        ("2020-W53-4T23:30:00-02:00", "2021-01-01T00:00:00Z", "2021-01-01T02:00:00Z"),
        ("2020-W53-5T00:30:00+02:00", "2020-12-31T22:00:00Z", "2020-12-31T23:00:00Z"),
        ("2020-W53-7T12:00:00.000001Z", "2021-01-03T12:00:00.000001Z", "2021-01-03T12:00:00.000001Z"),
    ]):
        yield {"kind": "fast", "input": text, "file": "", "sha256": "", "audit_id": f"timestamp-{index}",
               "options": {"min": earliest, "max": latest, "fast": True, "original": False}}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--kind", action="append", help="read-only check of selected existing case kinds")
    parser.add_argument("--output", type=Path, default=ROOT / "test-files/python-reference.json")
    parser.add_argument("--local-worker", choices=("windows_utc", "windows_eastern"))
    args = parser.parse_args()
    if args.kind and not args.check:
        parser.error("--kind requires --check")
    timezone = "EST5EDT" if args.local_worker == "windows_eastern" else "UTC"
    if os.environ.get("TZ") != timezone:
        raise SystemExit(subprocess.run([sys.executable, *sys.argv], env=os.environ | {"TZ": timezone}).returncode)
    if platform.python_implementation() != "CPython" or platform.python_version() != "3.14.6":
        raise RuntimeError("expected CPython 3.14.6")
    requirements = dict(line.split("==") for line in Path(__file__).with_name("requirements.txt").read_text().splitlines() if line)
    versions = {package: importlib.metadata.version(package) for package in requirements}
    if versions != requirements:
        raise RuntimeError(f"unexpected Python dependency versions: {versions}")

    import htmldate
    from htmldate import find_date
    from htmldate.extractors import custom_parse, extract_url_date, regex_parse, try_date_expr
    from htmldate.utils import Extractor
    from htmldate.validators import is_valid_date, validate_and_convert
    from lxml import etree

    hashes = {path.name: hashlib.sha256(path.read_bytes().replace(b"\r\n", b"\n")).hexdigest()
              for path in sorted(Path(htmldate.__file__).parent.glob("*.py"))}
    if hashes != SOURCES:
        raise RuntimeError("installed HtmlDate sources differ from the pinned upstream commit")
    logging.disable(logging.CRITICAL)
    warnings.simplefilter("ignore")
    environment = configure_reference(args.local_worker or "utc")
    if args.local_worker:
        print(json.dumps(local_cases(args.local_worker, environment), ensure_ascii=True))
        return
    if datetime(2024, 1, 1).timestamp() != 1704067200:
        raise RuntimeError("reference generation requires UTC C datetime timestamps")
    fixture = json.loads(args.output.read_text(encoding="utf-8"))
    originals = [case for case in fixture["cases"] if "audit_id" not in case]
    if len(originals) != 9533:
        raise RuntimeError("historical reference input inventory changed")
    previous = {case["audit_id"]: case for case in fixture["cases"] if "audit_id" in case}
    corpus_hashes = dict(fixture.get("corpus_sha256_lf", {}))
    candidates = fixture["cases"] if args.kind else originals + list(shortcut_cases())
    results = []
    previous_file, content = None, None
    for index, case in enumerate(candidates):
        kind = case["kind"]
        if args.kind and (kind not in args.kind or case.get("environment")):
            continue
        options = case["options"]
        earliest = datetime.fromisoformat(options["min"])
        latest = datetime.fromisoformat(options["max"])
        extensive = not options["fast"]
        text = case["input"]
        if kind == "file":
            if case["file"] != previous_file:
                content = (ROOT / case["file"]).read_bytes()
                previous_file = case["file"]
            normalized = content.replace(b"\r\n", b"\n")
            digest = hashlib.sha256(normalized).hexdigest()
            if case["file"] in corpus_hashes:
                if digest != corpus_hashes[case["file"]]:
                    raise RuntimeError(f"normalized corpus file changed: {case['file']}")
            elif hashlib.sha256(content).hexdigest() != case["sha256"]:
                raise RuntimeError(f"original corpus file changed: {case['file']}")
            corpus_hashes[case["file"]] = digest
            text = normalized
        is_valid_date.cache_clear()
        try_date_expr.cache_clear()
        error = ""
        try:
            if kind in ("html", "file"):
                value = find_date(text, extensive_search=extensive, original_date=options["original"], min_date=earliest, max_date=latest,
                                  url=options.get("url") or None, deferred_url_extractor=options.get("defer", False))
            elif kind == "fast":
                value = custom_parse(text, "%Y-%m-%d", earliest, latest)
            elif kind == "try":
                value = try_date_expr(text, "%Y-%m-%d", extensive, earliest, latest)
            elif kind == "regex":
                value = validate_and_convert(regex_parse(text), "%Y-%m-%d", earliest, latest)
            elif kind == "url":
                value = extract_url_date(text, Extractor(extensive, latest, earliest, options["original"], "%Y-%m-%d"))
            else:
                raise RuntimeError(f"unsupported fixture kind: {kind}")
        except Exception as exception:
            value, error = None, f"{type(exception).__name__}: {exception}"
        expected = {"date": value or "", "error": error}
        old = case if "expected" in case else previous.get(case.get("audit_id"))
        if old is not None and old["expected"] != expected:
            raise RuntimeError(f"Python result changed at {index}: {case.get('file') or case['input']!r}; {old['expected']} -> {expected}")
        results.append(case | {"expected": expected})
    if args.kind:
        print(f"Verified {len(results)} existing Python cases without rewriting the fixture")
        return
    for context in ("windows_utc", "windows_eastern"):
        encoded = subprocess.check_output([sys.executable, __file__, "--local-worker", context],
                                          env=os.environ | {"TZ": "EST5EDT" if context == "windows_eastern" else "UTC"})
        for case in json.loads(encoded):
            if case["audit_id"] in previous and previous[case["audit_id"]] != case:
                raise RuntimeError(f"local context result changed: {case['audit_id']}")
            results.append(case)
    document = {"python": platform.python_version(), "dependencies": versions, "libxml": ".".join(map(str, etree.LIBXML_VERSION)),
                "source_commit": UPSTREAM, "source_sha256": hashes, "source_line_endings": "LF",
                "current_time": FROZEN.isoformat(), "timezone_context": environment,
                "corpus_sha256_lf": dict(sorted(corpus_hashes.items())),
                "counts": Counter(case["kind"] for case in results), "cases": results,
                "go_only_cases": fixture["go_only_cases"]}
    encoded = (json.dumps(document, ensure_ascii=True, indent=2) + "\n").encode()
    if args.check:
        if args.output.read_bytes() != encoded:
            raise RuntimeError("Python reference bytes changed")
    else:
        args.output.write_bytes(encoded)
    print(f"Verified {len(originals)} unchanged historical results and {len(results) - len(originals)} additional Python cases")


if __name__ == "__main__":
    main()