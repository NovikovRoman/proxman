# PROXMAN

> A simple proxy list manager

[![Go Report Card](https://goreportcard.com/badge/github.com/NovikovRoman/proxman)](https://goreportcard.com/report/github.com/NovikovRoman/proxman)

## Usage example

```bash
go get github.com/NovikovRoman/proxman
```

```go
package main

import (
   "log"
   "math/rand/v2"
   "net/url"
   "sync"
   "time"

   "github.com/NovikovRoman/proxman"
)

func main() {
    p1, _ := url.Parse("http://127.0.0.1:8080")
    p2, _ := url.Parse("http://127.0.0.1:8081")
    p3, _ := url.Parse("http://127.0.0.2:8080")
    p4, _ := url.Parse("http://127.0.0.2:8081")
    p5, _ := url.Parse("http://127.0.0.3:8080")
    pm := proxman.New(p1, p2, p3, p4, p5)

    wg := &sync.WaitGroup{}
    for i := 0; i < 10; i++ {
        wg.Add(1)

        go func(wg *sync.WaitGroup, i int, pm *proxman.List) {
            defer wg.Done()

        var p *url.URL
        defer func() {
            if p != nil {
                // A random release option. Choose as needed.
                if rand.IntN(100) > 40 {
                    pm.Release(p)
                    log.Println(i, "simple release")
                } else { // deferred release
                    pm.DeferRelease(p, time.Second*3)
                    log.Println(i, "defer release")
                }
            }
        }()

        for {
            // In the process of using a proxy, they may be marked by you as banned:
            // pm.Ban(proxy, "some reason")
            // Then a check is needed:
            // …
            // if pm.NumBanned() == pm.Len() {
            //     log.Println("All proxies banned")
            //     break
            // }
            // …

            if p = pm.Acquire(); p == nil {
                time.Sleep(time.Second / 10)
                continue
            }
            break
        }

        if p != nil {
            log.Println(i, p)
        }
        }(wg, i, pm)
    }

    wg.Wait()

    // To demonstrate the defer release.
    for pm.NumFree() < pm.Num() {
       log.Printf("busy %+v\n", pm.BusyList())
       time.Sleep(time.Second / 4)
    }
}
```
