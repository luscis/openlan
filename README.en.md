# Overview 
[![codecov](https://codecov.io/gh/luscis/openlan/branch/master/graph/badge.svg)](https://codecov.io/gh/luscis/openlan)
[![Go Report Card](https://goreportcard.com/badge/github.com/luscis/openlan)](https://goreportcard.com/report/luscis/openlan-go)
[![GPL 3.0 License](https://img.shields.io/badge/License-GPL%203.0-blue.svg)](LICENSE)

The OpenLAN project help you to build a local area network via the Internet.  

## Terminology

* OLSW: OpenLAN Switch
* OLAP: OpenLAN Access Point
* NAT: Network Address translation

## Branch Access

                                        OLSW(Central) - 10.1.2.10/24
                                                ^
                                                |   
                                              Wifi(DNAT)
                                                |
                                                |
                       ----------------------Internet-------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   Branch 1                 Branch 2                     Branch 3    
                       |                        |                           |
                      OLAP                      OLAP                         OLAP
                 10.1.2.11/24              10.1.2.12/24                  10.1.2.13/24

## Multiple Area
                
                   192.168.1.20/24                                 192.168.1.22/24
                         |                                                 |
                        OLAP ---- Wifi ---> OLSW(NanJing) <---- Wifi --- OLAP
                                                |
                                                |
                                             Internet 
                                                |
                                                |
                                           OLSW(ShangHai) - 192.168.1.10/24
                                                |
                       ------------------------------------------------------
                       ^                        ^                           ^
                       |                        |                           |
                   Office Wifi               Home Wifi                 Hotel Wifi     
                       |                        |                           |
                     OLAP                     OLAP                         OLAP
                 192.168.1.11/24           192.168.1.12/24              192.168.1.13/24
