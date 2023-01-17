package mongo

import (
	"github.com/awakari/subscriptions/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_decodeRawKiwi(t *testing.T) {
	cases := map[string]struct {
		in  bson.M
		out Kiwi
		err error
	}{
		"ok": {
			in: bson.M{
				"key":     "k0",
				"pattern": "p0",
			},
			out: Kiwi{
				Key:     "k0",
				Pattern: "p0",
			},
		},
		"fail - missing key attribute": {
			in: bson.M{
				"pattern": "p0",
			},
			err: storage.ErrInternal,
		},
		"fail - missing pattern attribute": {
			in: bson.M{
				"key": "k0",
			},
			err: storage.ErrInternal,
		},
		"fail - invalid key attribute type": {
			in: bson.M{
				"key":     2,
				"pattern": "p0",
			},
			err: storage.ErrInternal,
		},
		"fail - invalid pattern attribute type": {
			in: bson.M{
				"key": "k0",
				"pattern": bson.M{
					"val": "p0",
				},
			},
			err: storage.ErrInternal,
		},
	}
	//
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := decodeRawKiwi(c.in)
			assert.ErrorIs(t, err, c.err)
			if c.err == nil {
				assert.Equal(t, c.out, out)
			}
		})
	}
}
