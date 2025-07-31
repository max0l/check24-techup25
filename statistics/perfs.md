# some statistics

Stats of 608e3ff6abd40990813ac2190e517c6bec581bf4
```
SVs: Beendet

 Performance counter stats for './readCSVs' (5 runs):

      207402832866      cycles:u                                                                ( +-  3.90% )  (83.30%)
      495025176603      instructions:u                   #    2.39  insn per cycle              ( +-  3.16% )  (83.38%)
       10701371714      cache-references:u                                                      ( +-  4.44% )  (83.31%)
         888118019      cache-misses:u                   #    8.30% of all cache refs           ( +-  7.56% )  (83.31%)
       95760499391      branch-instructions:u                                                   ( +-  3.14% )  (83.31%)
         652963902      branch-misses:u                  #    0.68% of all branches             ( +-  5.22% )  (83.39%)
            168876      page-faults:u                                                           ( +- 10.95% )
                 0      context-switches:u                                                    

          60.06574 +- 0.00448 seconds time elapsed  ( +-  0.01% )

perf stat -e  -r 5 --timeout 60000 ./readCSVs  261.40s user 242.10s system 167% cpu 5:00.37 total                                                                                                                                              /5m0.4s
```
