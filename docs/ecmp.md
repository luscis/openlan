# Topology
```
                      Access1         Access2         Access3        10.16.10.x/24
                           |             |              |
                           +-------------+--------------+
                                         |
                                         |
                                      Switch(BJ)                     10.16.10.1/24
                                        | |
                                        | |
                           +------------+ +-------------+
                           |                            |
   10.16.10.3/24           |                            |            10.16.10.2/24
                        Switch(NJ)                  Switch(WH)    
   10.18.10.3/24           |                            |            10.18.10.2/24
                           |                            |
                           +------------+ +-------------+
                                        | |
                                        | |
                                     Switch(SZ)                       10.18.10.1/24
                                         |
                                         |
                           +-------------+--------------+             10.18.10.x/24
                           |                            |
                        Access6                       Access7
```

# Test
On Access6
```
    ping 10.16.10.1
```
On Access7 
```
    ping 10.16.10.1
```
