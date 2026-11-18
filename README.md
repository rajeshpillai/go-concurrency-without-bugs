# Go concurrency without bugs: races, deadlocks and leaks

The code from the video **"Go concurrency without bugs: races, deadlocks and leaks"**. A crawler visits a fake
300-page website (built into each program, so every run is the same) and breaks in three classic ways, each caught
by a tool. Every output in the video is in `out/`, recorded by `run.py`.

| Folder | What it shows |
|---|---|
| `site`, `links`, `watch` | the fake website, the link reader, and the goroutine counter / "what is stuck" watchdog |
| `c0` | one page at a time (15 s) |
| `c1` | a goroutine per link, a shared map: crashes ("concurrent map writes"), `-race` finds it |
| `c2` | the same with a mutex: correct and fast, but floods the site |
| `c3bug` | a worker pool that deadlocks (hangs; the watchdog shows `chan send`) |
| `deadlock` | the smallest deadlock: "all goroutines are asleep - deadlock!" |
| `c4` | the pool with a `select` coordinator; hangs on a page that never answers (`-trap`) |
| `c5` | `context` deadlines: gives up on the broken page, finishes |
| `c6` | polite: a shared ticker, 50 requests a second |
| `leak` | a goroutine leak and its fix; try `GOEXPERIMENT=goroutineleakprofile go run ./leak` (Go 1.26) |
| `pipeline` | stages, fan-out, fan-in, cancellation |
| `checker` | `errgroup` with `SetLimit`: the first error stops the rest |
| `semaphore`, `ctxtree`, `once`, `shutdown` | a buffered-channel semaphore, a context tree, `sync.OnceValue`, Ctrl+C |
| `closing`, `selects`, `counter`, `loopvar`, `loopvar_old`, `wgvet`, `retry` | closing channels, select timeouts and default, mutex vs atomic vs channel, the Go 1.22 loop variable change, `go vet`'s WaitGroup check, `testing/synctest` |

```bash
go run ./c1                # crashes
go run -race ./c2          # no races
go run ./c4 -trap          # hangs (the watchdog reports after 30 s; -stuck 5s for less)
go run ./c5 -trap          # gives up on /page/137 after 1 s
go test -v ./retry         # 7 seconds of fake time, in no real time
```

Needs Go 1.26 (`WaitGroup.Go`, `testing/synctest`, the leak profile experiment). `run.py` re-runs everything.

## License

MIT
