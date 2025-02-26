package store

import (
	"context"
	"os"

	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/tracks"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const NIL_ID = -1

type DbLogMsg struct {
	Msg string `json:"msg" bson:"msg"`
}

type DbContent struct {
	Points []DbPoint  `json:"points"`
	Edges  []DbEdge   `json:"edges"`
	Tracks []DbTrack  `json:"tracks"`
	Log    []DbLogMsg `json:"log"`
}

type DbMeta struct {
	Tracks []DbTrack `json:"tracks"`
}

type DbItem struct {
	Id int64 `json:"id" bson:"id"`
}

type DbTrack struct {
	Id   int64            `json:"id" bson:"id"`
	Meta tracks.TrackMeta `json:"meta" bson:"meta"`
}

type GeoJsonGeometry struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}

type DbPoint struct {
	Id        int64           `json:"id" bson:"id"`
	Loc       GeoJsonGeometry `json:"loc" bson:"loc"`
	Tracks    []int64         `bson:"tracks"`
	Count     int             `bson:"count"`
	Begin     bool            `bson:"begin"`
	End       bool            `bson:"end"`
	Crossing  bool            `bson:"crossing"`
	Processed bool            `bson:"processed"`
}

type DbEdge struct {
	Id     string  `json:"id" bson:"id"`
	Points []int64 `json:"points" bson:"points"`
	Tracks []int64 `bson:"tracks"`
	Count  int     `bson:"count"`
}

type MongoStore struct {
	client      *mongo.Client
	db          *mongo.Database
	tracks      *mongo.Collection
	points      *mongo.Collection
	edges       *mongo.Collection
	tracksTmp   *mongo.Collection
	pointsTmp   *mongo.Collection
	edgesTmp    *mongo.Collection
	log         *mongo.Collection
	lastPointId int64
	lastTrackId int64
	cfg         *config.Configuration
}

func NewMongoStore(cfg *config.Configuration) *MongoStore {
	ms := MongoStore{}
	ms.cfg = cfg

	var err error

	uri := os.Getenv("MONGODB_URI")
	docs := "www.mongodb.com/docs/drivers/go/current/"
	if uri == "" {
		log.Error("Set your 'MONGODB_URI' environment variable. " +
			"See: " + docs +
			"usage-examples/#environment-variable")
		os.Exit(1)
	}

	log.Infof("connecting to %s", uri)

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

	ms.tracks = createCollection(ms.db, "tracks")
	ms.tracksTmp = createCollection(ms.db, "tracks_tmp")

	ms.points = createCollection(ms.db, "points")
	ms.pointsTmp = createCollection(ms.db, "points_tmp")

	ms.edges = createCollection(ms.db, "edges")
	ms.edgesTmp = createCollection(ms.db, "edges_tmp")

	ms.log = createCollection(ms.db, "log")

	// create indexes on points
	indexModels := []mongo.IndexModel{
		{Keys: bson.M{"loc": "2dsphere"}}, // geospatial index
		{Keys: bson.M{"id": 1}},
		{Keys: bson.M{"begin": 1}},
		{Keys: bson.M{"end": 1}},
		{Keys: bson.M{"crossing": 1}},
		{Keys: bson.M{"processed": 1}},
	}

	_, err = ms.points.Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		panic(err)
	}
	log.Debug("points index created")

	// create indexes on edges
	indexModels = []mongo.IndexModel{
		{Keys: bson.M{"id": 1}},
	}

	_, err = ms.edges.Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		panic(err)
	}
	log.Debug("edges index created")

	_, err = ms.tracks.Indexes().CreateMany(context.TODO(), indexModels)
	if err != nil {
		panic(err)
	}
	log.Debug("tracks index created")

	ms.lastTrackId = ms.GetMaxId(ms.tracks)
	ms.lastPointId = ms.GetMaxId(ms.points)

	log.Debugf("last geo id set to: %d", ms.lastPointId)

	return &ms
}

func (ms *MongoStore) Close() {
	log.Info("diconnecting from mongodb")
	if err := ms.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (ms *MongoStore) Log(msg string) {

	_, err := ms.log.InsertOne(context.TODO(), bson.M{"msg": msg})
	if err != nil {
		panic(err)
	}
}

func (ms *MongoStore) GetMaxId(collection *mongo.Collection) int64 {
	opts := options.Find().SetSort(bson.M{"id": NIL_ID}).SetLimit(1)
	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		panic(err)
	}
	var results []DbItem

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	if len(results) > 0 {
		return results[0].Id
	}

	return 0
}

func (ms *MongoStore) GenPointId() int64 {
	ms.lastPointId++
	return ms.lastPointId
}

func (ms *MongoStore) GenTrackId() int64 {
	ms.lastTrackId++
	return ms.lastTrackId
}

func (ms *MongoStore) Reset() {
	ms.tracks.DeleteMany(context.TODO(), bson.M{})
	ms.points.DeleteMany(context.TODO(), bson.M{})
	ms.edges.DeleteMany(context.TODO(), bson.M{})
	ms.log.DeleteMany(context.TODO(), bson.M{})
	ms.lastPointId = 0
	ms.lastTrackId = 0
}

func findPoint(id int64, points *[]DbPoint) *DbPoint {
	for i := 0; i < len(*points); i++ {
		if (*points)[i].Id == id {
			return &(*points)[i]
		}
	}
	return nil
}

func (ms *MongoStore) GetMeta() (DbMeta, error) {

	var meta DbMeta

	cursor, err := ms.tracks.Find(context.TODO(), bson.D{})
	if err != nil {
		return meta, err
	}

	if err = cursor.All(context.TODO(), &meta.Tracks); err != nil {
		return meta, err
	}

	return meta, nil
}

func (ms *MongoStore) GetLog() ([]DbLogMsg, error) {

	var result []DbLogMsg

	cursor, err := ms.log.Find(context.TODO(), bson.D{})
	if err != nil {
		return result, err
	}

	if err = cursor.All(context.TODO(), &result); err != nil {
		return result, err
	}

	return result, nil
}

func (ms *MongoStore) Export() *DbContent {

	var export DbContent

	cursor, err := ms.points.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &export.Points); err != nil {
		panic(err)
	}

	cursor, err = ms.edges.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &export.Edges); err != nil {
		panic(err)
	}

	cursor, err = ms.tracks.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &export.Tracks); err != nil {
		panic(err)
	}

	cursor, err = ms.log.Find(context.TODO(), bson.M{})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &export.Log); err != nil {
		panic(err)
	}

	return &export
}

func (ms *MongoStore) getPointsByIds(ids []int64, filter bson.M) []DbPoint {
	var points []DbPoint

	filter["id"] = bson.M{
		"$in": ids,
	}

	log.Debugf("filter=%v", filter)

	cursor, err := ms.points.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &points); err != nil {
		panic(err)
	}

	/*
		if len(ids) != len(points) {
			panic(fmt.Errorf("inconsistent number of ids (%d) and fetched entries (%d)", len(ids), len(points)))
		}
	*/

	return points
}

func (ms *MongoStore) updateSinglePoint(id int64, update bson.M) {

	_, err := ms.points.UpdateOne(context.TODO(), bson.M{"id": id}, bson.M{"$set": update})
	if err != nil {
		panic(err)
	}
}

func (ms *MongoStore) Import(in *DbContent) {

	ms.Reset()

	pointsI := make([]interface{}, len(in.Points))
	for i, v := range in.Points {
		pointsI[i] = v
	}

	_, err := ms.points.InsertMany(context.TODO(), pointsI)
	if err != nil {
		panic(err)
	}

	edgesI := make([]interface{}, len(in.Edges))
	for i, v := range in.Edges {
		edgesI[i] = v
	}

	_, err = ms.edges.InsertMany(context.TODO(), edgesI)
	if err != nil {
		panic(err)
	}

	tracksI := make([]interface{}, len(in.Tracks))
	for i, v := range in.Tracks {
		tracksI[i] = v
	}

	_, err = ms.tracks.InsertMany(context.TODO(), tracksI)
	if err != nil {
		panic(err)
	}

	logI := make([]interface{}, len(in.Log))
	for i, v := range in.Log {
		logI[i] = v
	}

	_, err = ms.log.InsertMany(context.TODO(), logI)
	if err != nil {
		panic(err)
	}

}

func (ms *MongoStore) StashPush() {
	_, err := ms.points.Aggregate(context.TODO(), bson.A{bson.M{"$out": "points_tmp"}})
	if err != nil {
		panic(err)
	}
	_, err = ms.edges.Aggregate(context.TODO(), bson.A{bson.M{"$out": "edges_tmp"}})
	if err != nil {
		panic(err)
	}
	_, err = ms.tracks.Aggregate(context.TODO(), bson.A{bson.M{"$out": "tracks_tmp"}})
	if err != nil {
		panic(err)
	}
}

func (ms *MongoStore) StashPop() {
	_, err := ms.pointsTmp.Aggregate(context.TODO(), bson.A{bson.M{"$out": "points"}})
	if err != nil {
		panic(err)
	}
	_, err = ms.edgesTmp.Aggregate(context.TODO(), bson.A{bson.M{"$out": "edges"}})
	if err != nil {
		panic(err)
	}
	_, err = ms.tracksTmp.Aggregate(context.TODO(), bson.A{bson.M{"$out": "tracks"}})
	if err != nil {
		panic(err)
	}
}

func createCollection(db *mongo.Database, name string) *mongo.Collection {
	err := db.CreateCollection(context.TODO(), name)
	if err != nil {
		panic(err)
	}
	return db.Collection(name)
}
