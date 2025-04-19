#!/usr/bin/env bash

go build

SVGPARAMS='--export --export-format=svg --points'

# unprocessed raw tracks
./geonet tracks test_data/t*.gpx $SVGPARAMS > doc/images/overview_tracks.svg

# generated network
./geonet net test_data/t* --interpolate --simplify --sim-min-dist 10 $SVGPARAMS > doc/images/overview_geonet.svg

# matching
./geonet tracks test_data/t1* test_data/t2* --interpolate $SVGPARAMS > doc/images/matching_tracks.svg
./geonet net test_data/t1* test_data/t2* --interpolate --match-max-dist 1 $SVGPARAMS > doc/images/matching_1.svg
./geonet net test_data/t1* test_data/t2* --interpolate --match-max-dist 5 $SVGPARAMS > doc/images/matching_5.svg
./geonet net test_data/t1* test_data/t2* --interpolate --match-max-dist 10 $SVGPARAMS > doc/images/matching_10.svg
./geonet net test_data/t1* test_data/t2* --interpolate --match-max-dist 50 $SVGPARAMS > doc/images/matching_50.svg

# interpolation
./geonet tracks test_data/t1* --export --export-format=svg --points --svg-edge-labels=false > doc/images/interpolation_track.svg
./geonet tracks test_data/t1* --interpolate --int-dist 10 $SVGPARAMS --svg-edge-labels=false > doc/images/interpolation_10.svg
./geonet tracks test_data/t1* --interpolate --int-dist 30 $SVGPARAMS --svg-edge-labels=false > doc/images/interpolation_30.svg

# simplification
./geonet tracks test_data/t4* $SVGPARAMS --svg-edge-labels=false > doc/images/simplify_track.svg
./geonet net test_data/t4* --match-max-dist 1 --simplify --sim-min-dist 10 $SVGPARAMS --svg-edge-labels=false > doc/images/simplify_10.svg
./geonet net test_data/t4* --match-max-dist 1 --simplify --sim-min-dist 50 $SVGPARAMS --svg-edge-labels=false > doc/images/simplify_50.svg
