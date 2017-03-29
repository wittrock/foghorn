# Foghorn

## Finding a good tuner gain value

```
~/src/rtl-ais(master*) » rtl_test -p
Found 1 device(s):
  0:  Realtek, RTL2838UHIDIR, SN: 00000001

Using device 0: Generic RTL2832U OEM
Found Rafael Micro R820T tuner
Supported gain values (29): 0.0 0.9 1.4 2.7 3.7 7.7 8.7 12.5 14.4 15.7 16.6 19.7 20.7 22.9 25.4 28.0 29.7 32.8 33.8 36.4 37.2 38.6 40.2 42.1 43.4 43.9 44.5 48.0 49.6
[R82XX] PLL not locked!
Sampling at 2048000 S/s.
Reporting PPM error measurement every 10 seconds...
Press ^C after a few minutes.
Reading samples in async mode...
real sample rate: 2047919 current PPM: -39 cumulative PPM: -39
real sample rate: 2047983 current PPM: -8 cumulative PPM: -23
real sample rate: 2048003 current PPM: 2 cumulative PPM: -14
real sample rate: 2047992 current PPM: -4 cumulative PPM: -12
real sample rate: 2047992 current PPM: -4 cumulative PPM: -10
real sample rate: 2047998 current PPM: -1 cumulative PPM: -9
real sample rate: 2048021 current PPM: 11 cumulative PPM: -6
real sample rate: 2047972 current PPM: -13 cumulative PPM: -7
real sample rate: 2047990 current PPM: -5 cumulative PPM: -7
```