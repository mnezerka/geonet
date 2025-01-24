package store

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"mnezerka/geonet/log"
	"mnezerka/geonet/tracks"

	"github.com/tkrajina/gpxgo/gpx"
	"go.mongodb.org/mongo-driver/bson"
)

func (ms *MongoStore) AddGpx(points []gpx.GPXPoint, meta tracks.TrackMeta) error {

	log.Infof("adding %s to the mongodb store", meta.Title)

	var lastPointId int64 = NIL_ID

	newTrack := DbTrack{
		Id:   ms.GenTrackId(),
		Meta: meta,
	}

	_, err := ms.tracks.InsertOne(context.TODO(), newTrack)
	if err != nil {
		return err
	}

	log.Debugf("registered track %d", newTrack.Id)

	for i := 0; i < len(points); i++ {
		point := points[i]
		lastPointId, err = ms.AddGpxPoint(
			&point,
			newTrack.Id,        // id of the current track
			lastPointId,        // id of the previous point
			i == 0,             // is point beginning of the track?
			i == len(points)-1, // is point end of the track?
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *MongoStore) AddGpxPoint(point *gpx.GPXPoint, trackId int64, lastPointId int64, begin, end bool) (int64, error) {

	var finalPointId int64

	cntr := bson.M{
		"type":        "Point",
		"coordinates": []float64{point.Longitude, point.Latitude},
	}

	// look for existing point to be reused
	cursor, err := ms.points.Find(
		context.TODO(),
		bson.M{
			"loc": bson.M{
				"$nearSphere": bson.M{
					"$geometry":    cntr,
					"$maxDistance": ms.cfg.MatchMaxDistance, // distance in meters
				},
			},
		})

	if err != nil {
		return 0, err
	}

	var nearest []DbPoint

	err = cursor.All(context.TODO(), &nearest)
	if err != nil {
		return 0, err
	}

	var nps []string
	for _, p := range nearest {
		nps = append(nps, fmt.Sprintf("%d", p.Id))
	}
	log.Debugf("nearest points: %s", strings.Join(nps, ","))

	// if there is some point close to the actual one
	if len(nearest) > 0 {
		// get first (nearest) one
		n := nearest[0]

		log.Debugf("reusing point %v", n)

		// increase counter
		//update := bson.M{"$set": bson.M{"count": n.Count + 1}}
		update := bson.M{
			"count": n.Count + 1,
			"begin": n.Begin || begin,
			"end":   n.End || end,
		}

		// if current track is not associated with reused point, add it
		if !slices.Contains(n.Tracks, trackId) {
			update["tracks"] = append(n.Tracks, trackId)
		}

		log.Debugf("updating point, id: %d update: %v", n.Id, update)

		_, err = ms.points.UpdateOne(context.TODO(), bson.M{"id": n.Id}, bson.M{"$set": update})
		if err != nil {
			return 0, err
		}

		finalPointId = n.Id
	} else {
		// no near point exists, register new one

		//logging.debug('registring new point %s track_id=%s (%s)', final_point_id, track_id, point)
		newPoint := DbPoint{
			Id: ms.GenPointId(),
			Loc: GeoJsonGeometry{
				Type:        "Point",
				Coordinates: []float64{point.Longitude, point.Latitude},
			},
			Tracks: []int64{trackId},
			Count:  1,
			Begin:  begin,
			End:    end,
		}

		log.Debugf("creating new point %v", newPoint)

		_, err = ms.points.InsertOne(context.TODO(), newPoint)
		if err != nil {
			return 0, err
		}

		finalPointId = newPoint.Id
	}

	// --------------------  edge processing
	if lastPointId != NIL_ID {

		// ignore self edges
		if lastPointId != finalPointId {
			// create edge with sorted point ids to avoid duplicates (reverse direction of track movement)
			edgePoints := []int64{min(lastPointId, finalPointId), max(lastPointId, finalPointId)}
			edgeId := fmt.Sprintf("%d-%d", edgePoints[0], edgePoints[1])

			log.Debugf("new edge %s", edgeId)

			newEdge := DbEdge{
				Id:     edgeId,
				Points: edgePoints,
				Tracks: []int64{trackId},
				Count:  1,
			}

			// check if edge already exists
			var existingEdges []DbEdge
			cursor, err = ms.edges.Find(context.TODO(), bson.M{"id": edgeId})
			if err != nil {
				return NIL_ID, err
			}

			if err = cursor.All(context.TODO(), &existingEdges); err != nil {
				return NIL_ID, err
			}

			if len(existingEdges) > 0 {

				existingEdge := existingEdges[0]
				log.Debugf("reusing existing edge: %s", existingEdge.Id)
				update := bson.M{
					"count": existingEdge.Count + 1,
				}

				// if current track is not associated with reused edge, add it
				if !slices.Contains(existingEdge.Tracks, trackId) {
					update["tracks"] = append(existingEdge.Tracks, trackId)
				}

				log.Debugf("updating %s %v", existingEdge.Id, update)

				_, err = ms.edges.UpdateOne(context.TODO(), bson.M{"id": existingEdge.Id}, bson.M{"$set": update})
				if err != nil {
					return 0, err
				}
			} else {
				log.Debugf("registering new edge: %s", newEdge.Id)
				_, err = ms.edges.InsertOne(context.TODO(), newEdge)
				if err != nil {
					return NIL_ID, err
				}

				// update crossings
				err = ms.updateEdgeCrossings(&newEdge, newEdge.Points[0])
				if err != nil {
					return NIL_ID, err
				}
				err = ms.updateEdgeCrossings(&newEdge, newEdge.Points[1])
				if err != nil {
					return NIL_ID, err
				}
			}
		}
	}

	return finalPointId, nil
}

func (ms *MongoStore) updateEdgeCrossings(edge *DbEdge, pointId int64) error {

	log.Debugf("updating crossings for point %d and edge %s", pointId, edge.Id)

	// check edges for point 1
	cursor, err := ms.edges.Find(context.TODO(), bson.M{
		"$or": bson.A{
			bson.M{"points.0": pointId},
			bson.M{"points.1": pointId},
		},
	})
	if err != nil {
		return err
	}

	var edges []DbEdge

	if err = cursor.All(context.TODO(), &edges); err != nil {
		return err
	}

	log.Debugf("edges for point %d: %v", pointId, edges)

	if len(edges) > 2 {
		log.Debugf("crossings detected (point %d and edge %s), updating point", pointId, edge.Id)
		update := bson.M{
			"crossing": true,
		}

		log.Debugf("mongo update - point: %d, update: %v", pointId, update)
		_, err = ms.points.UpdateOne(context.TODO(), bson.M{"id": pointId}, bson.M{"$set": update})
		if err != nil {
			return err
		}
	}

	return nil
}
