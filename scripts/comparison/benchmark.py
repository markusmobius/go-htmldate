"""Compare immutable Go-HtmlDate revisions with serial, paired warm corpus passes."""

import argparse
from datetime import datetime, timezone
import hashlib
import io
import json
import math
import os
from pathlib import Path
import platform
import signal
import statistics
import subprocess
import tarfile
import tempfile
import time


ROOT = Path(__file__).resolve().parents[2]
REFERENCES = {
    "v1.9.3": "79c39d1b1e4240383f82d140f3d79788df949ecb",
    "v1.10.1": "0f04a39fb476a75744ed948bfcd887306ca3f187",
}
COHORTS = tuple(
    f"document-{date}-{mode}"
    for date in ("original", "modified")
    for mode in ("fast", "extensive")
)


def capture(command, cwd, environment):
    return subprocess.check_output(command, cwd=cwd, env=environment, text=True, timeout=60)


def module_graph(source, environment):
    remaining = capture(["go", "list", "-m", "-json", "all"], source, environment).strip()
    modules = []
    decoder = json.JSONDecoder()
    while remaining:
        module, offset = decoder.raw_decode(remaining)
        if module.get("Replace"):
            raise RuntimeError("Benchmark dependencies must not use replacements")
        modules.append({key: module[key] for key in (
            "Path", "Version", "Main", "GoVersion", "Sum", "GoModSum"
        ) if key in module})
        remaining = remaining[offset:].lstrip()
    return modules


def build(version, commit, work, environment, git):
    resolved = capture([git, "rev-parse", f"{commit}^{{commit}}"], ROOT, environment).strip()
    if resolved != commit:
        raise RuntimeError(f"Unexpected source identity for {version}: {resolved}")
    source = work / version
    source.mkdir()
    archive = subprocess.check_output([
        git, "archive", "--format=tar", commit, "--", "go.mod", "go.sum", "*.go", "internal",
    ], cwd=ROOT, env=environment, timeout=60)
    with tarfile.open(fileobj=io.BytesIO(archive)) as snapshot:
        snapshot.extractall(source, filter="data")
    runner = Path("scripts/comparison")
    (source / runner).mkdir(parents=True, exist_ok=True)
    overlay = work / f"{version}-overlay.json"
    replacements = {
        str(source / runner / path.name): str(path)
        for path in sorted((ROOT / runner).glob("*.go"))
    }
    overlay.write_text(json.dumps({"Replace": replacements}), encoding="utf-8")
    binary = work / f"htmldate-{version}"
    subprocess.run([
        "go", "build", "-mod=readonly", "-trimpath", "-overlay", str(overlay),
        "-o", str(binary), "./scripts/comparison",
    ], cwd=source, env=environment, check=True, timeout=60)
    modules = module_graph(source, environment)
    subprocess.run(["go", "mod", "verify"], cwd=source, env=environment, check=True, timeout=60)
    return binary, {
        "commit": commit,
        "source_archive_sha256": hashlib.sha256(archive).hexdigest(),
        "go_mod_sha256": hashlib.sha256((source / "go.mod").read_bytes()).hexdigest(),
        "go_sum_sha256": hashlib.sha256((source / "go.sum").read_bytes()).hexdigest(),
        "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
        "build_info": capture(["go", "version", "-m", str(binary)], source, environment),
        "modules": modules,
    }


def timeout_expired(signal_number, frame):
    raise TimeoutError("Benchmark exceeded its 60-second operation or six-minute overall limit")


def read_message(process, deadline):
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise TimeoutError("Benchmark exceeded its six-minute overall limit")
    signal.setitimer(signal.ITIMER_REAL, min(60, remaining))
    try:
        line = process.stdout.readline()
        if not line:
            raise RuntimeError(f"Benchmark process exited unexpectedly: {process.poll()}")
        return json.loads(line)
    finally:
        signal.setitimer(signal.ITIMER_REAL, 0)


def measure(process, cohort, warmup, deadline):
    process.stdin.write(json.dumps({"cohort": cohort, "warmup": warmup}) + "\n")
    process.stdin.flush()
    result = read_message(process, deadline)
    metadata = result["metadata"]
    if metadata["cohort"] != cohort or metadata["cases"] != 1000 or result["warmup"] != warmup:
        raise RuntimeError(f"Unexpected benchmark response: {cohort}")
    if not warmup and (not math.isfinite(result["pass_ms"]) or result["pass_ms"] <= 0):
        raise RuntimeError("Invalid benchmark timings")
    if warmup and (len(result["results"]) != 1000 or any(outcome["error"] for outcome in result["results"])):
        raise RuntimeError(f"Extraction errors in {cohort}")
    return result


def save_report(path, report):
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    temporary.replace(path)


def summarize(samples):
    summary = {}
    before, after = REFERENCES
    for cohort in sorted({sample["cohort"] for sample in samples}):
        selected = [sample for sample in samples if sample["cohort"] == cohort]
        measurements = {}
        timings = {}
        for version in REFERENCES:
            records = [sample for sample in selected if sample["version"] == version]
            timings[version] = [record["pass_ms"] for record in records]
            median = statistics.median(timings[version])
            cases = records[0]["metadata"]["cases"]
            measurements[version] = {
                "median_ms": median,
                "range_ms": [min(timings[version]), max(timings[version])],
                "samples_ms": timings[version],
                "ms_per_page": median / cases,
                "pages_per_second": cases * 1000 / median,
                "bytes_per_page": statistics.median(
                    record["allocated_bytes"] / cases for record in records
                ),
                "allocations_per_page": statistics.median(
                    record["allocations"] / cases for record in records
                ),
            }
        measurements["old_over_new_time"] = measurements[before]["median_ms"] / measurements[after]["median_ms"]
        ratios = [old / new for old, new in zip(timings[before], timings[after], strict=True)]
        measurements["paired_old_over_new"] = {
            "median": statistics.median(ratios), "range": [min(ratios), max(ratios)], "samples": ratios,
        }
        summary[cohort] = measurements
    return summary


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runs", type=int, default=8, help="single timed corpus passes per version and mode")
    parser.add_argument("--cpu", type=int, default=2)
    parser.add_argument("--cohort", action="append", choices=COHORTS)
    parser.add_argument("--min-date", default="1995-01-01")
    parser.add_argument("--max-date", default="2026-09-13")
    parser.add_argument("--current-time", default="2026-09-13T12:00:00Z")
    parser.add_argument("--git", default="git")
    parser.add_argument("--output", type=Path, required=True)
    arguments = parser.parse_args()
    if arguments.runs < 2 or arguments.runs % 2:
        parser.error("use an even number of runs >= 2")
    if not hasattr(os, "sched_getaffinity") or arguments.cpu not in os.sched_getaffinity(0):
        parser.error("run under Linux/WSL and select an available CPU")
    cohorts = list(arguments.cohort or COHORTS)
    if len(set(cohorts)) != len(cohorts):
        parser.error("select each cohort at most once")
    if arguments.output.exists():
        parser.error("output already exists; choose a new report path")
    started = time.monotonic()
    deadline = started + 360
    signal.signal(signal.SIGALRM, timeout_expired)
    signal.signal(signal.SIGTERM, timeout_expired)
    environment = os.environ.copy()
    settings = {
        "GOTOOLCHAIN": "go1.27.1", "GOWORK": "off", "GOOS": "linux", "GOARCH": "amd64",
        "CGO_ENABLED": "0", "GOAMD64": "v1", "GOMAXPROCS": "1", "TZ": "UTC",
        "GOGC": "100", "GOMEMLIMIT": "off", "GOFLAGS": "", "GOEXPERIMENT": "", "GODEBUG": "",
        "LC_ALL": "C.UTF-8",
    }
    environment.update(settings)
    goroot = capture(["go", "env", "GOROOT"], ROOT, environment).strip()
    environment["ZONEINFO"] = str(Path(goroot) / "lib/time/zoneinfo.zip")
    runner_files = sorted((ROOT / "scripts/comparison").glob("*.go")) + [Path(__file__).resolve()]
    report = {
        "measured_at_utc": datetime.now(timezone.utc).isoformat(),
        "platform": platform.platform(),
        "cpu": next(line.split(":", 1)[1].strip() for line in Path("/proc/cpuinfo").read_text().splitlines()
                    if line.startswith("model name")),
        "cpu_affinity": [arguments.cpu], "environment": settings,
        "zoneinfo_sha256": hashlib.sha256(Path(environment["ZONEINFO"]).read_bytes()).hexdigest(),
        "status": "in-progress",
        "runs_per_version_and_cohort": arguments.runs,
        "total_timed_passes": len(cohorts) * len(REFERENCES) * arguments.runs,
        "processes": "One persistent process per version; corpus loaded and parsed once per process.",
        "warmup": "One untimed full pass per version/mode before all timed passes; eight warmups total.",
        "timing": "Serial extraction loop only; loading, initial DOM parse, validation and result formatting excluded.",
        "corpus_normalization": "CRLF to LF before hashing, parsing and timing, identically for both versions.",
        "source_references": {}, "runner_sha256": {
            path.relative_to(ROOT).as_posix(): hashlib.sha256(path.read_bytes().replace(b"\r\n", b"\n")).hexdigest()
            for path in runner_files
        },
        "fixtures": [], "outcomes": {}, "warmup_metadata": {}, "samples": [],
    }
    save_report(arguments.output, report)
    with tempfile.TemporaryDirectory(prefix="htmldate-benchmark-") as directory:
        work = Path(directory)
        binaries = {}
        for version, commit in REFERENCES.items():
            print(f"Build {version} ({commit[:12]})", flush=True)
            binaries[version], report["source_references"][version] = build(
                version, commit, work, environment, arguments.git,
            )
            save_report(arguments.output, report)
        workers = {}
        try:
            for version, binary in binaries.items():
                print(f"Load and parse {version} corpus once", flush=True)
                command = [
                    "taskset", "-c", str(arguments.cpu), str(binary), "-benchmark", "document", "-passes", "1",
                    "-corpus-root", str(ROOT), "-min-date", arguments.min_date,
                    "-max-date", arguments.max_date, "-current-time", arguments.current_time,
                ]
                workers[version] = subprocess.Popen(
                    command, cwd=ROOT, env=environment, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                    text=True, encoding="utf-8", bufsize=1,
                )
                ready = read_message(workers[version], deadline)
                if ready["metadata"]["cases"] != 1000 or (report["fixtures"] and ready["fixtures"] != report["fixtures"]):
                    raise RuntimeError("Versions received different corpus bytes")
                report["fixtures"] = ready["fixtures"]
            for cohort in cohorts:
                for version, worker in workers.items():
                    print(f"Warmup {version} {cohort}", flush=True)
                    result = measure(worker, cohort, True, deadline)
                    report["outcomes"][f"{version}/{cohort}"] = result["results"]
                    report["warmup_metadata"][f"{version}/{cohort}"] = result["metadata"]
                    save_report(arguments.output, report)
            for round_index in range(arguments.runs):
                ordered_cohorts = cohorts[round_index % len(cohorts):] + cohorts[:round_index % len(cohorts)]
                if round_index % 2:
                    ordered_cohorts.reverse()
                versions = list(binaries)
                if round_index % 2:
                    versions.reverse()
                for cohort in ordered_cohorts:
                    for version in versions:
                        result = measure(workers[version], cohort, False, deadline)
                        if result["metadata"] != report["warmup_metadata"][f"{version}/{cohort}"]:
                            raise RuntimeError(f"Outputs or settings changed: {version} {cohort}")
                        result.update(round=round_index + 1, version=version, cohort=cohort)
                        report["samples"].append(result)
                        save_report(arguments.output, report)
                        print(f"{len(report['samples'])}/{report['total_timed_passes']} {version} {cohort}: "
                              f"{result['pass_ms']:.2f} ms/pass", flush=True)
        except BaseException as error:
            report.update(status="incomplete", error=str(error))
            save_report(arguments.output, report)
            raise
        finally:
            for worker in workers.values():
                worker.stdin.close()
            for worker in workers.values():
                try:
                    worker.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    worker.kill()
                    worker.wait()
    report["summary"] = summarize(report["samples"])
    before, after = REFERENCES
    report["differences"] = {
        cohort: [
            {"file": old["file"], before: old["date"], after: new["date"]}
            for old, new in zip(report["outcomes"][f"{before}/{cohort}"],
                                report["outcomes"][f"{after}/{cohort}"], strict=True)
            if old != new
        ] for cohort in cohorts
    }
    report.update(status="complete", elapsed_seconds=time.monotonic() - started)
    save_report(arguments.output, report)
    print("Mode | v1.9.3 ms/pass | v1.10.1 ms/pass | Old/New", flush=True)
    for cohort, result in report["summary"].items():
        print(f"{cohort} | {result[before]['median_ms']:.2f} | {result[after]['median_ms']:.2f} | "
              f"{result['old_over_new_time']:.3f}x", flush=True)
    print(f"Completed {len(report['samples'])} timed passes in {report['elapsed_seconds']:.1f} seconds", flush=True)
    print(f"Report: {arguments.output}", flush=True)


if __name__ == "__main__":
    main()