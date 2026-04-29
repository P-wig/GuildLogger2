package db

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ToObjectID(id string) (primitive.ObjectID, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid object id %q: %w", id, err)
	}
	return oid, nil
}

func SerializeDoc(doc bson.M) bson.M {
	out := make(bson.M, len(doc))
	for k, v := range doc {
		out[k] = v
	}

	if id, ok := out["_id"].(primitive.ObjectID); ok {
		out["_id"] = id.Hex()
	}
	return out
}

func SerializeDocs(docs []bson.M) []bson.M {
	out := make([]bson.M, 0, len(docs))
	for _, d := range docs {
		out = append(out, SerializeDoc(d))
	}
	return out
}
