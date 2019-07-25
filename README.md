# kiwisdr-qrss-grabber
Grab QRSS images from KiwiSDRs.

### grabloop.go

Loop forever, grabbing QRSS audio from KiwiSDRs
and creating images in a spool directory.

### render.go

Render a 16-bit little-endian mono audio file,
either raw (like created by go-kiwisdr-client)
or a simple Riff (like created by kiwiclient.y),
as an FFT image.
