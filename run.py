"""Run every demo shown in "Go concurrency without bugs" and record what it printed.

    python3 code/run.py          # everything (a few minutes; waits for a quiet machine)

Each demo writes code/out/<name>.txt: the command as typed ("$ ..."), then the real output.
The crawlers run against a fake site inside the same program (code/site), so every run visits the
same 300 pages. Timings vary a little run to run; the narration reads these files.
"""
import os
import re
import subprocess
import sys
import time
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE / "out"
BIN = HERE / ".bin"


def quiet(limit=2.0):
    waited = 0
    while os.getloadavg()[0] > limit:
        if waited % 300 == 0:
            print(f"waiting for a quiet machine (load {os.getloadavg()[0]:.1f})", flush=True)
        time.sleep(15)
        waited += 15


def clean(s):
    s = s.replace(str(HERE) + "/", "./").replace(str(HERE), ".")
    s = re.sub(r"http://127\.0\.0\.1:\d+", "http://127.0.0.1:PORT", s)   # the fake site's random port
    s = re.sub(r"0x[0-9a-f]+", "0x…", s)                                  # addresses change every run
    return s.rstrip("\n")


def demo(name, *cmds, shown=None, env=None):
    quiet()
    pairs = []
    for i, c in enumerate(cmds):
        r = subprocess.run(["bash", "-c", c], cwd=HERE, capture_output=True, text=True,
                           env={**os.environ, **(env or {})})
        pairs.append((shown[i] if shown and shown[i] else c, clean(r.stdout + r.stderr)))
    txt = "\n".join(f"$ {c}\n{o}" if o else f"$ {c}" for c, o in pairs)
    OUT.mkdir(parents=True, exist_ok=True)
    (OUT / f"{name}.txt").write_text(txt + "\n")
    print(f"--- {name}\n{txt}", flush=True)


def build():
    BIN.mkdir(exist_ok=True)
    for p in ["c0", "c1", "c2", "c3bug", "c4", "c5", "c6", "deadlock", "leak", "checker"]:
        subprocess.run(["go", "build", "-o", str(BIN / p), f"./{p}"], cwd=HERE, check=True)
    subprocess.run(["go", "build", "-race", "-o", str(BIN / "c1-race"), "./c1"], cwd=HERE, check=True)
    subprocess.run(["go", "build", "-race", "-o", str(BIN / "c2-race"), "./c2"], cwd=HERE, check=True)
    subprocess.run(["go", "build", "-o", str(BIN / "leak-profile"), "./leak"], cwd=HERE, check=True,
                   env={**os.environ, "GOEXPERIMENT": "goroutineleakprofile"})


def main():
    build()
    demo("c0", ".bin/c0", shown=["go run ./c0"])
    demo("c1", ".bin/c1 2>&1 | grep 'fatal error'", ".bin/c1 2>&1 | grep 'fatal error'", ".bin/c1 2>&1 | grep 'fatal error'",
         shown=["go run ./c1", "go run ./c1", "go run ./c1"])
    demo("c1_race", ".bin/c1-race 2>&1 | grep -c 'WARNING: DATA RACE'",
         ".bin/c1-race 2>&1 | sed -n '/WARNING: DATA RACE/,/^==================/p' | sed -n '1,/^==================/p' | grep -v '^=================='",
         shown=["go run -race ./c1 2>&1 | grep -c 'WARNING: DATA RACE'", "go run -race ./c1     # the first report, in full"])
    demo("c2", ".bin/c2", ".bin/c2-race", shown=["go run ./c2", "go run -race ./c2"])
    demo("deadlock", ".bin/deadlock 2>&1 | head -4", shown=["go run ./deadlock"])
    demo("c3bug", ".bin/c3bug", shown=["go run ./c3bug"])
    demo("c4", ".bin/c4", shown=["go run ./c4"])
    demo("c4_trap", ".bin/c4 -trap -stuck 5s", shown=["go run ./c4 -trap -stuck 5s"])
    demo("c5_trap", ".bin/c5 -trap", shown=["go run ./c5 -trap"])
    for w in [1, 2, 4, 8, 16, 32, 64]:
        demo(f"workers_{w}", f".bin/c5 -workers {w}", shown=[f"go run ./c5 -workers {w}"])
    demo("checker", ".bin/checker", ".bin/checker -trap", shown=["go run ./checker", "go run ./checker -trap"])
    demo("leak", ".bin/leak", ".bin/leak -fixed", shown=["go run ./leak", "go run ./leak -fixed"])
    demo("leak_profile", ".bin/leak-profile | head -4", shown=["GOEXPERIMENT=goroutineleakprofile go run ./leak"])
    demo("retry", "go test -v -count=1 ./retry 2>&1 | grep -v '^=== RUN'", shown=["go test -v ./retry"])
    demo("c6", ".bin/c6", ".bin/c6 -workers 32", shown=["go run ./c6", "go run ./c6 -workers 32"])


def extra():
    for p in ["closing", "pipeline", "selects", "loopvar"]:
        subprocess.run(["go", "build", "-o", str(BIN / p), f"./{p}"], cwd=HERE, check=True)
    demo("closing", ".bin/closing", ".bin/closing send 2>&1 | grep -m1 panic",
         shown=["go run ./closing", "go run ./closing send"])
    demo("counter", "go test -run x -bench . -count 3 ./counter | grep Bench",
         shown=["go test -bench . -count 3 ./counter"])
    demo("pipeline", ".bin/pipeline", shown=["go run ./pipeline"])
    demo("wgvet", "go vet ./wgvet 2>&1 | grep -v '^#'", shown=["go vet ./wgvet"])
    demo("selects", ".bin/selects", shown=["go run ./selects"])
    demo("loopvar", ".bin/loopvar", "cd loopvar_old && GOTOOLCHAIN=local go run .",
         shown=["go run ./loopvar          # go.mod says go 1.26", "go run ./loopvar_old      # go.mod says go 1.21"])


def extra2():
    for p in ["shutdown", "semaphore", "ctxtree", "once"]:
        subprocess.run(["go", "build", "-o", str(BIN / p), f"./{p}"], cwd=HERE, check=True)
    demo("semaphore", ".bin/semaphore", shown=["go run ./semaphore"])
    demo("ctxtree", ".bin/ctxtree", shown=["go run ./ctxtree"])
    demo("once", ".bin/once", shown=["go run ./once"])
    demo("shutdown", ".bin/shutdown & P=$!; sleep 2; kill -INT $P; wait $P",
         shown=["go run ./shutdown     # then, 2 seconds later, Ctrl+C"])


if __name__ == "__main__":
    if sys.argv[1:] == ["extra"]:
        extra()
    elif sys.argv[1:] == ["extra2"]:
        extra2()
    else:
        main()
        extra()
        extra2()
