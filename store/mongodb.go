package store

import (
	"context"
	"fmt"
	"os"
	"slices"

	"mnezerka/geonet/log"

	gj "github.com/paulmach/go.geojson"

	"github.com/tkrajina/gpxgo/gpx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MAX_SPOT_DISTANCE = 75

type TrackMeta struct {
	Name     string `bson:"name"`
	FilePath string `bson:"filepath"`
}

type GeoJsonGeometry struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type DbPoint struct {
	Id     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Loc    GeoJsonGeometry      `json:"loc" bson:"loc"`
	Tracks []primitive.ObjectID `bson:"tracks"`
	Count  int                  `bson:"count"`
}

type MongoStore struct {
	client *mongo.Client
	db     *mongo.Database
	tracks *mongo.Collection
	points *mongo.Collection
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

	// create geospatial index
	indexModel := mongo.IndexModel{
		Keys: bson.M{"loc": "2dsphere"},
	}
	log.Debug("index created")

	_, err = ms.points.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		panic(err)
	}

	return &ms
}

func (ms *MongoStore) Close() {
	log.Info("diconnecting from mongodb")
	if err := ms.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (ms *MongoStore) Reset() {
	ms.tracks.DeleteMany(context.TODO(), bson.M{})
	ms.points.DeleteMany(context.TODO(), bson.M{})
}

func (ms *MongoStore) AddGpx(gpx *gpx.GPX, filePath string) error {

	log.Infof("adding %s to the mongodb store", filePath)

	lastPointId := primitive.NilObjectID

	for i := 0; i < len(gpx.Tracks); i++ {
		track := gpx.Tracks[i]

		newTrackMeta := TrackMeta{Name: fmt.Sprintf("%s-%d", gpx.Name, i), FilePath: filePath}
		result, err := ms.tracks.InsertOne(context.TODO(), newTrackMeta)
		if err != nil {
			return err
		}

		log.Debugf("%v", result.InsertedID)

		for j := 0; j < len(track.Segments); j++ {
			segment := track.Segments[j]
			for k := 0; k < len(segment.Points); k++ {
				point := segment.Points[k]
				lastPointId, err = ms.AddGpxPoint(&point, result.InsertedID.(primitive.ObjectID), nil)
				if err != nil {
					return err
				}

				log.Infof("added, id=%v", lastPointId.Hex())

				/*
					if k > 2 {
						break
					}
				*/
			}
		}
	}

	// last_point_id = None
	// for point in points:
	//
	//	   last_point_id = self.add_point(point, track_id, last_point_id)
	//	}
	return nil
}

func (ms *MongoStore) AddGpxPoint(point *gpx.GPXPoint, trackId primitive.ObjectID, lastPoint *gpx.GPXPoint) (primitive.ObjectID, error) {
	//logging.debug('add point: %s, track_id=%s, last_point_id: %s', point, track_id, last_point_id if last_point_id is not None else "-")

	var finalPointId primitive.ObjectID

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
		return primitive.NilObjectID, err
	}

	var nearest []DbPoint

	err = cursor.All(context.TODO(), &nearest)
	if err != nil {
		return primitive.NilObjectID, err
	}

	log.Debugf("Nearest points: %v", nearest)

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

		log.Debugf("updating %v %v", n.Id, update)

		_, err = ms.points.UpdateOne(context.TODO(), bson.M{"_id": n.Id}, bson.M{"$set": update})
		if err != nil {
			return primitive.NilObjectID, err
		}

		log.Debug("updated")

		finalPointId = n.Id
	} else {
		// no near point exists, register new one

		//logging.debug('registring new point %s track_id=%s (%s)', final_point_id, track_id, point)
		newPoint := DbPoint{
			Loc: GeoJsonGeometry{
				Type:        "Point",
				Coordinates: []float64{point.Longitude, point.Latitude},
			},
			Tracks: []primitive.ObjectID{trackId},
			Count:  1,
		}

		log.Infof("creating new point %v", newPoint)

		result, err := ms.points.InsertOne(context.TODO(), newPoint)
		if err != nil {
			return primitive.NilObjectID, err
		}

		log.Infof("point stored %v", result.InsertedID)

		finalPointId = result.InsertedID.(primitive.ObjectID)

		/*
		   # --------------------  edge processing
		   if last_point_id is not None:

		       # ignore self edges
		       if last_point_id != final_point_id:
		           # create edge with sorted point ids to avoid duplicates (reverse direction of track movement)
		           edge_points = (last_point_id, final_point_id) if last_point_id < final_point_id else (final_point_id, last_point_id)
		           edge_id = f'{edge_points[0]}-{edge_points[1]}'
		           edge = {
		               'index': edge_id,
		               'p1': edge_points[0],
		               'p2': edge_points[1],
		               'tracks': [track_id],
		               'q': 1
		           }

		           existing = self.db.edges.find_one({'index': edge_id})

		           if existing is not None:
		               logging.debug('reusing existing edge: %s', existing['index'])

		               update = {
		                   'q': loc['q'] + 1
		               }

		               if track_id not in existing['tracks']:
		                   update['tracks'] = existing['tracks'].copy()
		                   update['tracks'].append(track_id)

		               self.db.edges.update_one({'_id': existing['_id']}, {'$set': update})

		           else:
		               logging.debug('registering new edge: %s', edge_id)
		               self.db.edges.insert_one(edge)

		   return final_point_id
		*/
	}
	return finalPointId, nil
}

func (ms *MongoStore) ToGeoJson() (*gj.FeatureCollection, error) {

	// geos = []
	// index = {}
	//var geos []interface{}

	collection := gj.NewFeatureCollection()

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

		pnt := gj.NewPointFeature(point.Loc.Coordinates)
		collection.AddFeature(pnt)
	}

	// render edges
	/*
	   if show_edges:
	       edges = self.db.edges.find({})

	       for edge in edges:
	           p1 = index[edge['p1']]    # p1 index -> p1 data
	           p2 = index[edge['p2']]    # p2 index -> p2 data
	           line = {
	               'type': 'Feature',
	               'properties': {
	                   'edge': edge['index'],
	                   'tracks': edge['tracks'],
	                   'q': edge['q']
	               },
	               'geometry': {
	                   'coordinates': [p1['loc']['coordinates'], p2['loc']['coordinates']],
	                   'type': 'LineString'
	               }
	           }
	           geos.append(line)
	*/

	/*
		geometries := GeoJsonFeatureCollection{
			Type:     "FeatureCollection",
			Features: geos,
		}
	*/

	return collection, nil
}
