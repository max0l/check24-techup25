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

Stats of 4a194fd5ed95442ef4a8a2b9b78913ddff64afe5 (with 10 workes)
```
 Performance counter stats for './readCSVs' (5 runs):

      284431718560      cycles:u                                                                ( +-  0.64% )  (83.33%)
      611394345876      instructions:u                   #    2.15  insn per cycle              ( +-  0.59% )  (83.40%)
       14006941250      cache-references:u                                                      ( +-  0.68% )  (83.30%)
        1553121940      cache-misses:u                   #   11.09% of all cache refs           ( +-  0.93% )  (83.37%)
      118071454171      branch-instructions:u                                                   ( +-  0.60% )  (83.29%)
         946593376      branch-misses:u                  #    0.80% of all branches             ( +-  0.59% )  (83.31%)
            380858      page-faults:u                                                           ( +- 15.97% )
                 0      context-switches:u                                                    

          60.07289 +- 0.00164 seconds time elapsed  ( +-  0.00% )

perf stat -e  -r 5 --timeout 60000 ./readCSVs  361.49s user 275.17s system 211% cpu 5:00.39 total 
```