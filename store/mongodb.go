package store

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"mnezerka/geonet/log"

	geojson "github.com/paulmach/go.geojson"

	"github.com/tkrajina/gpxgo/gpx"
	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MAX_SPOT_DISTANCE = 75
const NIL_GEO_ID = -1

type TrackMeta struct {
	GeoId    int64  `json:"geoid" bson:"geoid"`
	Name     string `bson:"name"`
	FilePath string `bson:"filepath"`
}

type GeoJsonGeometry struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type DbPoint struct {
	GeoId  int64           `json:"geoid" bson:"geoid"`
	Loc    GeoJsonGeometry `json:"loc" bson:"loc"`
	Tracks []int64         `bson:"tracks"`
	Count  int             `bson:"count"`
}

type DbEdge struct {
	Id     string  `json:"id" bson:"id"`
	Points []int64 `json:"points" bson:"points"`
	Tracks []int64 `bson:"tracks"`
	Count  int     `bson:"count"`
}

type MongoStore struct {
	client    *mongo.Client
	db        *mongo.Database
	tracks    *mongo.Collection
	points    *mongo.Collection
	edges     *mongo.Collection
	lastGeoId int64
}

func NewMongoStore() *MongoStore {
	ms := MongoStore{}

	var err error

	uri := os.Getenv("MONGODB_URI")
	log.Infof("connecting to %s", uri)
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {

		log.Debug("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
		os.Exit(1)
	}

	ms.client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	log.Info("connected to mongodb")

	err = ms.client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	log.Info("ping mongodb passed")

	ms.db = ms.client.Database("geonet")

	err = ms.db.CreateCollection(context.TODO(), "tracks")
	if err != nil {
		panic(err)
	}
	ms.tracks = ms.db.Collection("tracks")

	err = ms.db.CreateCollection(context.TODO(), "points")
	if err != nil {
		panic(err)
	}
	ms.points = ms.db.Collection("points")

	err = ms.db.CreateCollection(context.TODO(), "edges")
	if err != nil {
		panic(err)
	}
	ms.edges = ms.db.Collection("edges")

	// create indexes on points
	indexModels := []mongo.IndexModel{
		{Keys: bson.M{"loc": "2dsphere"}}, // geospatial index
		{Keys: bson.M{"geoid": 1}},
	}

	_, err = ms.points.Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		panic(err)
	}
	log.Debug("points index created")

	ms.lastGeoId = ms.GetMaxGeoId()

	log.Debugf("last geo id set to: %d", ms.lastGeoId)

	return &ms
}

func (ms *MongoStore) Close() {
	log.Info("diconnecting from mongodb")
	if err := ms.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (ms *MongoStore) GetMaxGeoId() int64 {
	opts := options.Find().SetSort(bson.M{"geoid": NIL_GEO_ID}).SetLimit(1)
	cursor, err := ms.points.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		panic(err)
	}
	var results []DbPoint

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	if len(results) > 0 {
		return results[0].GeoId
	}

	return 0
}

func (ms *MongoStore) GenGeoId() int64 {
	ms.lastGeoId++
	return ms.lastGeoId
}

func (ms *MongoStore) Reset() {
	ms.tracks.DeleteMany(context.TODO(), bson.M{})
	ms.points.DeleteMany(context.TODO(), bson.M{})
	ms.edges.DeleteMany(context.TODO(), bson.M{})
	ms.lastGeoId = 0
}

func (ms *MongoStore) AddGpx(points []gpx.GPXPoint, filePath string) error {

	log.Infof("adding %s to the mongodb store", filePath)

	var lastPointId int64 = NIL_GEO_ID

	newTrackMeta := TrackMeta{
		GeoId:    ms.GenGeoId(),
		Name:     filePath,
		FilePath: filePath,
	}

	_, err := ms.tracks.InsertOne(context.TODO(), newTrackMeta)
	if err != nil {
		return err
	}

	log.Debugf("registered track %d", newTrackMeta.GeoId)

	for i := 0; i < len(points); i++ {
		point := points[i]
		lastPointId, err = ms.AddGpxPoint(&point, newTrackMeta.GeoId, lastPointId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ms *MongoStore) AddGpxPoint(point *gpx.GPXPoint, trackId int64, lastPointId int64) (int64, error) {

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
					"$maxDistance": MAX_SPOT_DISTANCE, // distance in meters
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
		nps = append(nps, fmt.Sprintf("%d", p.GeoId))
	}
	log.Debugf("Nearest points: %s", strings.Join(nps, ","))

	// if there is some point close to the actual one
	if len(nearest) > 0 {
		// get first (nearest) one
		n := nearest[0]

		log.Debugf("reusing point %v", n)

		// increase counter
		//update := bson.M{"$set": bson.M{"count": n.Count + 1}}
		update := bson.M{"count": n.Count + 1}

		// if current track is not associated with reused point, add it
		if !slices.Contains(n.Tracks, trackId) {
			update["tracks"] = append(n.Tracks, trackId)
		}

		log.Debugf("updating point, geoid: %d update: %v", n.GeoId, update)

		_, err = ms.points.UpdateOne(context.TODO(), bson.M{"geoid": n.GeoId}, bson.M{"$set": update})
		if err != nil {
			return 0, err
		}

		finalPointId = n.GeoId
	} else {
		// no near point exists, register new one

		//logging.debug('registring new point %s track_id=%s (%s)', final_point_id, track_id, point)
		newPoint := DbPoint{
			GeoId: ms.GenGeoId(),
			Loc: GeoJsonGeometry{
				Type:        "Point",
				Coordinates: []float64{point.Longitude, point.Latitude},
			},
			Tracks: []int64{trackId},
			Count:  1,
		}

		log.Debugf("creating new point %v", newPoint)

		_, err := ms.points.InsertOne(context.TODO(), newPoint)
		if err != nil {
			return 0, err
		}

		finalPointId = newPoint.GeoId
	}

	// --------------------  edge processing
	if lastPointId != NIL_GEO_ID {

		// ignore self edges
		if lastPointId != finalPointId {
			// create edge with sorted point ids to avoid duplicates (reverse direction of track movement)
			edgePoints := []int64{min(lastPointId), max(finalPointId)}
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
				return NIL_GEO_ID, err
			}

			if err = cursor.All(context.TODO(), &existingEdges); err != nil {
				return NIL_GEO_ID, err
			}

			if len(existingEdges) > 0 {

				existingEdge := existingEdges[0]
				log.Debugf("reusing existing edge: %s", existingEdge.Id)
				update := bson.M{
					"count": existingEdge.Count + 1,
				}

				// if current track is not associated with reused point, add it
				if !slices.Contains(existingEdge.Tracks, trackId) {
					update["tracks"] = append(existingEdge.Tracks, trackId)
				}

				log.Debugf("updating %d %v", existingEdge.Id, update)

				_, err = ms.edges.UpdateOne(context.TODO(), bson.M{"d": existingEdge.Id}, bson.M{"$set": update})
				if err != nil {
					return 0, err
				}
			} else {
				log.Debugf("registering new edge: %s", newEdge.Id)
				_, err = ms.edges.InsertOne(context.TODO(), newEdge)
				if err != nil {
					return NIL_GEO_ID, err
				}

			}
		}
	}

	return finalPointId, nil
}

func (ms *MongoStore) ToGeoJson() (*geojson.FeatureCollection, error) {

	collection := geojson.NewFeatureCollection()

	// render points and build dict of points for searching when
	// rendering edges

	cursor, err := ms.points.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	var points []DbPoint

	err = cursor.All(context.TODO(), &points)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(points); i++ {
		point := points[i]

		//pnt := geojson.NewFeature(orb.Point{point.Loc.Coordinates[0], point.Loc.Coordinates[1]})
		pnt := geojson.NewPointFeature(point.Loc.Coordinates)

		collection.AddFeature(pnt)
	}

	// render edges
	cursor, err = ms.edges.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	var edges []DbEdge

	err = cursor.All(context.TODO(), &edges)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(edges); i++ {
		edge := edges[i]
		p1 := findPoint(edge.Points[0], &points)
		p2 := findPoint(edge.Points[1], &points)

		// of some of the points were not found -> database is inconsistent
		if p1 == nil || p2 == nil {
			return nil, fmt.Errorf("inconsistent database, some edge points where not found %d %d", edge.Points[0], edge.Points[1])
		}

		log.Debugf("edge: %v, p1: %v p2: %v ", edge, p1, p2)

		edgeCoordinates := [][]float64{p1.Loc.Coordinates, p2.Loc.Coordinates}

		line := geojson.NewLineStringFeature(edgeCoordinates)
		collection.AddFeature(line)
	}

	return collection, nil
}

func findPoint(geoId int64, points *[]DbPoint) *DbPoint {
	for i := 0; i < len(*points); i++ {
		if (*points)[i].GeoId == geoId {
			return &(*points)[i]
		}
	}
	return nil
}
