"""Check the current scenario using immutable Tools policy and catalog inputs."""
import hashlib
import json
from pathlib import Path
import stat
import sys
import time
import types
import urllib.request


ROOT = Path(__file__).resolve().parents[2]
HELPER_SOURCE = "0e8ff6fee42e88858069a6f6ae6f1080d95c5eae"
HELPER_URL = (
    "https://raw.githubusercontent.com/faustbrian/go-library-tools/"
    + HELPER_SOURCE + "/tools/compatibility/generate.py"
)
HELPER_SHA256 = "fb3585b3e539cc179183e5fd431d3871a17d4960aa54eb9c0c0815e2eb08c6b1"
ASSET_BASE = "https://github.com/faustbrian/go-library-tools/releases/download/v1.8.4/"
ASSETS = {
    "compatibility-sets.json": "bc80555c2ac0d61010e83848f6e8f649a057624124a32bb4bab2fedbee229740",
    "catalog-consumer.json": "e7ae64cf3f1a9d7c14771b99d12438a061b38ec1d56099c54adedc58ffce167d",
    "catalog-engineering.json": "835088c5b7f70a1bf350bf9a811abe3dc710bbd82abc599b456eed0cfbb86883",
}
TOOLS_HEADER = "module github.com/faustbrian/go-library-tools/release/compatibility-consumer\n"
ASSURANCE_HEADER = "module github.com/faustbrian/go-library-assurance/scenarios/compatibility-consumer\n"


def fetch_verified(url, digest, limit, *, open_url=None, deadline=None):
    # Keep each response private until its full digest passes. The injected opener
    # models transport for refusal tests; callers own its trust and timing behavior.
    # No credentials, ambient proxy credentials, mutable refs or cached fallback.
    deadline = time.monotonic() + 90 if deadline is None else deadline
    remaining = deadline - time.monotonic()
    if remaining <= 0:
        raise TimeoutError("bootstrap time budget exceeded")
    if open_url is None:
        class BoundedRedirect(urllib.request.HTTPRedirectHandler):
            max_redirections = 3
            max_repeats = 1

            def redirect_request(self, request, response, code, message, headers, new_url):
                if time.monotonic() >= deadline:
                    raise TimeoutError("bootstrap time budget exceeded")
                if not new_url.startswith("https://"):
                    raise ValueError("bootstrap redirect requires HTTPS")
                return super().redirect_request(request, response, code, message, headers, new_url)

        open_url = urllib.request.build_opener(
            urllib.request.ProxyHandler({}), BoundedRedirect(),
        ).open
    chunks = []
    total = 0
    with open_url(url, timeout=min(15, remaining)) as response:
        read = getattr(response, "read1", response.read)
        while True:
            if time.monotonic() >= deadline:
                raise TimeoutError("bootstrap time budget exceeded")
            chunk = read(min(65536, limit + 1 - total))
            if not chunk:
                break
            total += len(chunk)
            if total > limit:
                raise ValueError("bootstrap byte budget exceeded")
            chunks.append(chunk)
    body = b"".join(chunks)
    if hashlib.sha256(body).hexdigest() != digest:
        raise ValueError("bootstrap digest mismatch")
    return body


def decode_json(body):
    def unique_pairs(pairs):
        value = {}
        for key, item in pairs:
            if key in value:
                raise ValueError("duplicate JSON key")
            value[key] = item
        return value

    return json.loads(body, object_pairs_hook=unique_pairs)


def load_inputs():
    # All downloads share one budget. A helper with correct syntax but the wrong
    # bytes cannot reach compile/exec, and catalog assets cannot fall back to cache.
    deadline = time.monotonic() + 90
    source = fetch_verified(HELPER_URL, HELPER_SHA256, 1024 * 1024, deadline=deadline)
    assets = {
        name: decode_json(fetch_verified(
            ASSET_BASE + name, digest, 16 * 1024 * 1024, deadline=deadline,
        ))
        for name, digest in ASSETS.items()
    }
    helper = types.ModuleType("verified_tools_compatibility")
    helper.__file__ = str(ROOT / "tools/compatibility/generate.py")
    # Only digest-verified published code executes. Invoke its pure functions;
    # do not call main(), change its defaults or maintain another policy engine.
    exec(compile(source, HELPER_URL, "exec"), helper.__dict__)
    catalogs = {}
    for name in ("catalog-consumer.json", "catalog-engineering.json"):
        for module in assets[name]["modules"]:
            catalogs.setdefault(module["module_path"], {}).update(module)
    return helper, assets["compatibility-sets.json"], list(catalogs.values())


def render_selection(inputs, selection):
    # Only the module identity is adapted. Imports remain exact renderer bytes;
    # adding a manual comment or allowing a second selection source would drift.
    helper, sets, catalogs = inputs
    selected, catalogs = helper.current_consumer_inputs(
        sets, catalogs, selection, verify_public=False,
    )
    generated, imports = helper.render_clean_consumer(
        selected, catalogs, generator_path="../../tools/compatibility/check-selection.py",
    )
    if not generated.startswith(TOOLS_HEADER):
        raise ValueError("published helper module header changed")
    generated = ASSURANCE_HEADER + generated[len(TOOLS_HEADER):]
    return len(selected["modules"]), generated, imports


def validate_selection(inputs, selection, go_mod, test_source):
    count, generated, imports = render_selection(inputs, selection)
    if not inputs[0].clean_consumer_go_mod_matches(go_mod, generated):
        raise ValueError("current scenario is stale: go.mod")
    if test_source != imports:
        raise ValueError("current scenario is stale: consumer_test.go")
    return count


def read_owned(name, limit):
    # Admit bounded regular input from a stable checkout. These checks precede
    # open but do not provide an atomic snapshot against concurrent replacement.
    if name not in ("selection.json", "go.mod", "consumer_test.go"):
        raise ValueError("unknown scenario input")
    directory = ROOT / "scenarios/compatibility-consumer"
    for parent in (ROOT / "scenarios", directory):
        if not stat.S_ISDIR(parent.lstat().st_mode):
            raise ValueError("unsafe scenario directory")
    path = directory / name
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_nlink != 1 or info.st_size > limit:
        raise ValueError("missing or unsafe scenario input")
    with path.open("rb") as handle:
        body = handle.read(limit + 1)
    if len(body) > limit:
        raise ValueError("scenario input exceeds byte budget")
    return body


def maintain_scenario(inputs, *, write=False):
    # Validate every bounded owned input and direct dependency before any write.
    # --write replaces only the generated import source, never requirements,
    # selection or frozen inputs. A stable checkout is required during this call.
    selection = decode_json(read_owned("selection.json", 65536))
    go_mod = read_owned("go.mod", 1024 * 1024).decode()
    test_source = read_owned("consumer_test.go", 1024 * 1024).decode()
    count, generated, imports = render_selection(inputs, selection)
    if not inputs[0].clean_consumer_go_mod_matches(go_mod, generated):
        raise ValueError("current scenario is stale: go.mod")
    if write:
        (ROOT / "scenarios/compatibility-consumer/consumer_test.go").write_text(
            imports, encoding="utf-8",
        )
    elif test_source != imports:
        raise ValueError("current scenario is stale: consumer_test.go")
    return count


def main(arguments=None):
    arguments = sys.argv[1:] if arguments is None else arguments
    if arguments not in ([], ["--write"]):
        raise ValueError("usage: python3 tools/compatibility/check-selection.py [--write]")
    write = arguments == ["--write"]
    count = maintain_scenario(load_inputs(), write=write)
    operation = "regenerated imports for" if write else "selection matches:"
    print(f"current scenario {operation} {count} public modules")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, KeyError, TypeError) as error:
        print(f"selection check failed: {error}", file=sys.stderr)
        sys.exit(1)
