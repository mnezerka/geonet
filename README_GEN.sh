#!/usr/bin/env bash

go build

# unprocessed raw tracks
./geonet tracks test_data/t*.gpx --export --export-format svg --points > doc/images/overview_tracks.svg

# generated network
./geonet gen test_data/t* --interpolate --simplify --sim-min-dist 10 --export --export-format=svg --points > doc/images/overview_geonet.svg

# matching
./geonet tracks test_data/t1* test_data/t2* --interpolate --export --export-format=svg --points > doc/images/matching_tracks.svg
./geonet gen test_data/t1* test_data/t2* --interpolate --match-max-dist 1 --export --export-format=svg --points > doc/images/matching_1.svg
./geonet gen test_data/t1* test_data/t2* --interpolate --match-max-dist 5 --export --export-format=svg --points > doc/images/matching_5.svg
./geonet gen test_data/t1* test_data/t2* --interpolate --match-max-dist 10 --export --export-format=svg --points > doc/images/matching_10.svg
./geonet gen test_data/t1* test_data/t2* --interpolate --match-max-dist 50 --export --export-format=svg --points > doc/images/matching_50.svg

# interpolation
./geonet tracks test_data/t1* --export --export-format=svg --points > doc/images/interpolation_track.svg
./geonet tracks test_data/t1* --interpolate --int-dist 10 --export --export-format=svg --points > doc/images/interpolation_10.svg
./geonet tracks test_data/t1* --interpolate --int-dist 30 --export --export-format=svg --points > doc/images/interpolation_30.svg

# simplification
./geonet tracks test_data/t4* --export --export-format=svg --points > doc/images/simplify_track.svg
./geonet gen test_data/t4* --match-max-dist 1 --simplify --sim-min-dist 10 --export --export-format=svg --points > doc/images/simplify_10.svg
./geonet gen test_data/t4* --match-max-dist 1 --simplify --sim-min-dist 50 --export --export-format=svg --points > doc/images/simplify_50.svg
