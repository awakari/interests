package mongo

import (
	"fmt"
	"github.com/awakari/subscriptions/storage"
	"go.mongodb.org/mongo-driver/bson"
)

// Kiwi represents the key-pattern pair (or Key-Input WIldcard)
type Kiwi struct {
	Key     string `bson:"key"`
	Pattern string `bson:"pattern"`
}

const kiwiAttrKey = "key"
const kiwiAttrPattern = "pattern"

func decodeRawKiwi(raw bson.M) (k Kiwi, err error) {
	var present bool
	k.Key, present = raw[kiwiAttrKey].(string)
	if present {
		k.Pattern, present = raw[kiwiAttrPattern].(string)
	}
	if !present {
		err = fmt.Errorf("%w: failed to decode kiwi %v", storage.ErrInternal, raw)
	}
	return
}
