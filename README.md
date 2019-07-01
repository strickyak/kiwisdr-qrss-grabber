# kiwisdr-qrss-grabber
Grab QRSS images from KiwiSDRs.

## Requirements:

* Recent golang
* Python 2, probably 2.7
* numpy
* Works on my linux amd64.  Others I don't know.
* Install this as "$HOME/go/src/github.com/strickyak/kiwisdr-qrss-grabber/".
* You may need to export GOPATH="$HOME/go" if this is not the default for your golang.

## Limitations

* This code is brand new and only experimental quality.
* It needs tweaking and tuning all over.
* Lots of constants are hardwired.
* It grabs two or three times as much spectrum as most QRSS grabbers.
* That is useful when things are barely working, since it should grab WSPR as well.

## Configure:

Edit publish.sh to define how to publish images.

By default it will install them under $HOME/pub.qrss/.

## Run:

    $ cd "$HOME/go/src/github.com/strickyak/kiwisdr-qrss-grabber/"
    $
    $ bash run.sh W6REK-40 sybil.yak.net 7038.5 11

That will make 11-minute grabs from KiwiSDR host sybil.yak.net (on the default port 8073)
with the base frequency set to 7038.5 kHz (the actual grab will be 1 to 1.7 kHz above that)
and identify the grabs as "W6REK-40".  It repeats over and over again.
You can run more than one of these at a time, for different bands.

Unless you're doing quick one-off testing on from a non-busy site,
*PLEASE only do this when you have permission from the owner of the KiwiSDR!*
If you do not own the KiwiSDR, you do not automatically have the right to permanently tie up
a significant portion of its FPGA and internet bandwidth.
(If this becomes a problem, I'm worried more KiwiSDRs will add passwords.)

## Contact:

Email me with username "strick" in the domain "yak" with the TLD "net".
