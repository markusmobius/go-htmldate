"""Compare two Go revisions and two pinned Python htmldate checkouts offline."""

import argparse
from datetime import datetime, time
import hashlib
import importlib.metadata
import json
import logging
from pathlib import Path
import subprocess
import sys


MODES = {
    "Fast": (True, False),
    "Extensive": (True, True),
    "ModifiedFast": (False, False),
    "ModifiedExtensive": (False, True),
}
REPO = Path(__file__).resolve().parents[2]


def read_results(path):
    with path.open(encoding="utf-8-sig") as source:
        rows = [json.loads(line) for line in source if line.strip()]
    results = {row["URL"]: row for row in rows}
    if not results or len(results) != len(rows):
        raise ValueError(f"Empty results or duplicate URLs in {path}")
    return results


def fixture_path(filename):
    for directory in ("mediacloud", "comparison", "mock"):
        path = REPO / "test-files" / directory / filename
        if path.is_file():
            return path
    raise FileNotFoundError(filename)


def python_results(args):
    sys.path.insert(0, str(args.worker.resolve()))
    from htmldate import find_date

    logging.disable(logging.CRITICAL)
    min_date = datetime.fromisoformat(args.min_date)
    max_date = datetime.combine(datetime.fromisoformat(args.max_date).date(), time.max)
    entries = read_results(args.manifest)
    with args.output.open("w", encoding="utf-8") as destination:
        for index, entry in enumerate(entries.values(), 1):
            content = fixture_path(entry["File"]).read_bytes()
            result = {key: entry[key] for key in ("URL", "File")}
            for mode, (original, extensive) in MODES.items():
                result[mode] = find_date(
                    content,
                    original_date=original,
                    extensive_search=extensive,
                    min_date=min_date,
                    max_date=max_date,
                ) or ""
            destination.write(json.dumps(result, ensure_ascii=True) + "\n")
            if index % 100 == 0:
                print(f"{args.worker.name}: {index}/{len(entries)} pages", flush=True)


def git_revision(path):
    return subprocess.check_output(
        ["git", "-C", str(path), "rev-parse", "HEAD"], text=True
    ).strip()


def run_go(checkout, output, args, overlay=None):
    command = [args.go, "run", "-mod=mod"]
    if overlay:
        command += ["-overlay", str(overlay)]
    command += [
        "./scripts/comparison", "-json",
        "-min-date", args.min_date, "-max-date", args.max_date,
    ]
    print(f"Go: {checkout}", flush=True)
    with output.open("wb") as destination:
        subprocess.run(command, cwd=checkout, stdout=destination, check=True)


def report(args, runs):
    labels = {}
    for filename in ("eval_mediacloud_2020.json", "eval_default.json"):
        with (args.python_after / "tests" / filename).open(encoding="utf-8") as source:
            labels.update(json.load(source))

    entries = runs["go-before"]
    for name, results in runs.items():
        if results.keys() != entries.keys():
            raise ValueError(f"{name} does not cover the same URLs")
        if any(results[url]["File"] != entry["File"] for url, entry in entries.items()):
            raise ValueError(f"{name} does not use the same fixtures")
    if any(url not in labels for url in entries):
        raise ValueError("Some corpus URLs have no upstream evaluation entry")

    summary = {"documents": len(entries), "min_date": args.min_date, "max_date": args.max_date}
    summary["revisions"] = {
        "go-before": git_revision(args.go_before),
        "go-after-base": git_revision(REPO),
        "python-before": git_revision(args.python_before),
        "python-after": git_revision(args.python_after),
    }
    summary["python"] = sys.version
    summary["dependencies"] = {
        package: importlib.metadata.version(package)
        for package in ("dateparser", "python-dateutil", "lxml", "charset-normalizer", "urllib3", "regex")
    }
    summary.update(scores={}, agreement={}, changes={})
    details = {"changes": {}, "disagreements": {}}

    for name, results in runs.items():
        summary["scores"][name] = {}
        for mode in ("Fast", "Extensive"):
            counts = {"TP": 0, "FP": 0, "FN": 0, "TN": 0}
            for url, result in results.items():
                expected = labels[url]["date"]
                actual = result[mode]
                if not actual:
                    counts["TN" if expected is None or expected == "" else "FN"] += 1
                else:
                    counts["TP" if actual == expected else "FP"] += 1
            correct = counts["TP"] + counts["TN"]
            summary["scores"][name][mode] = dict(counts, accuracy=correct / len(entries))

    for stage in ("before", "after"):
        summary["agreement"][stage] = {}
        details["disagreements"][stage] = {}
        for mode in MODES:
            differences = [
                {"URL": url, "File": entry["File"], "Go": runs[f"go-{stage}"][url][mode],
                 "Python": runs[f"python-{stage}"][url][mode]}
                for url, entry in entries.items()
                if runs[f"go-{stage}"][url][mode] != runs[f"python-{stage}"][url][mode]
            ]
            summary["agreement"][stage][mode] = len(entries) - len(differences)
            details["disagreements"][stage][mode] = differences

    for language in ("go", "python"):
        summary["changes"][language] = {}
        details["changes"][language] = {}
        for mode in MODES:
            differences = [
                {"URL": url, "File": entry["File"], "Before": runs[f"{language}-before"][url][mode],
                 "After": runs[f"{language}-after"][url][mode]}
                for url, entry in entries.items()
                if runs[f"{language}-before"][url][mode] != runs[f"{language}-after"][url][mode]
            ]
            summary["changes"][language][mode] = len(differences)
            details["changes"][language][mode] = differences

    fixture_hashes = {
        entry["File"]: hashlib.sha256(fixture_path(entry["File"]).read_bytes()).hexdigest()
        for entry in entries.values()
    }
    for filename, value in (("summary.json", summary), ("differences.json", details), ("fixtures.json", fixture_hashes)):
        (args.output_dir / filename).write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, indent=2))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--go-before", type=Path)
    parser.add_argument("--python-before", type=Path)
    parser.add_argument("--python-after", type=Path)
    parser.add_argument("--output-dir", type=Path)
    parser.add_argument("--go", default="go")
    parser.add_argument(
        "--dateparser-version", default="1.4.3",
        help="Required Python dateparser reference version (default: %(default)s)",
    )
    parser.add_argument("--min-date", default="1995-01-01")
    parser.add_argument("--max-date", default=datetime.now().strftime("%Y-%m-%d"))
    parser.add_argument("--worker", type=Path, help=argparse.SUPPRESS)
    parser.add_argument("--manifest", type=Path, help=argparse.SUPPRESS)
    parser.add_argument("--output", type=Path, help=argparse.SUPPRESS)
    args = parser.parse_args()
    dateparser_version = importlib.metadata.version("dateparser")
    if dateparser_version != args.dateparser_version:
        parser.error(
            f"expected Python dateparser {args.dateparser_version}, found {dateparser_version}; "
            "install the pinned reference dependencies"
        )
    if args.worker:
        python_results(args)
        return
    if not all((args.go_before, args.python_before, args.python_after, args.output_dir)):
        parser.error("--go-before, --python-before, --python-after and --output-dir are required")
    args.output_dir = args.output_dir.resolve()
    args.output_dir.mkdir(parents=True, exist_ok=True)
    overlay = args.output_dir / "baseline-overlay.json"
    source = Path("scripts/comparison/main.go")
    overlay.write_text(json.dumps({"Replace": {
        str((args.go_before / source).resolve()): str(REPO / source),
    }}), encoding="utf-8")
    paths = {name: args.output_dir / f"{name}.jsonl" for name in ("go-before", "go-after", "python-before", "python-after")}
    run_go(args.go_before, paths["go-before"], args, overlay)
    run_go(REPO, paths["go-after"], args)
    for stage in ("before", "after"):
        subprocess.run([
            sys.executable, str(Path(__file__).resolve()),
            "--worker", str(getattr(args, f"python_{stage}")),
            "--manifest", str(paths["go-before"]), "--output", str(paths[f"python-{stage}"]),
            "--min-date", args.min_date, "--max-date", args.max_date,
            "--dateparser-version", args.dateparser_version,
        ], check=True)
    report(args, {name: read_results(path) for name, path in paths.items()})


if __name__ == "__main__":
    main()