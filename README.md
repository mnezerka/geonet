# GeoNet

GeoNet is a library for generating net from multiple gps tracks.

[![Go](https://github.com/mnezerka/geonet/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/mnezerka/geonet/actions/workflows/go.yml)

Current state of the algorithm could be checkd on [Cases page](https://mnezerka.github.io/geonet/)

## Usage

Generate geonet from gpx files in `data` directory. Save net in data.geonet
```bash
geonet gen data/*gpx --interpolate --simplify --save > data.geonet
```

Load net from file and generate html page:
```bash
geonet load data.geonet --export > data.json
```

## Algorithm

- tracks are added one by one
- for each track
  - for each point
    - look for existing point which is nearby
    - if such point exists, reuse it, else register as new point
    - create edge for each newly registered point
