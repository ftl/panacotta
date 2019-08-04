# panacotta

Panacotta is a panadapter for my FT-450D. It uses a RTL28xxU SDR dongle that is connected to the IF of the transceiver to generate a FFT and a waterfall around the currently tuned VFO frequency. The IF is tapped out using the PAT70 designed by G4HUP (sk) and sold by [SDR-Kits](https://sdr-kits.net/Panoramic-Adaptor-Tap-Boards).

I have plenty of ideas what should be added to such a display of the frequency domain in order to be useful for a radio amateur.

## Hint
This is all work in progress and currently not in a state that is useful for anybody. As soon as there is more value in this little tool, I will also provide more documentation about the inner workings. So stay tuned ;-)

## Build

Install `librtlsdr-dev`:
```
sudo apt install librtlsdr-dev
```

Build command (your GTK framework version may vary):
```
go build -tags gtk_3_22
```

## Disclaimer
I develop this tool for myself and just for fun in my free time. If you find it useful, I'm happy to hear about that. If you have trouble using it, you have all the source code to fix the problem yourself (although pull requests are welcome).

## License
This tool is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/) 2019