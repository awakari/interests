package mongo

// kiwi represents the key-pattern pair (or Key-Input WIldcard)
type kiwi struct {
	Key     string `bson:"key"`
	Pattern string `bson:"pattern"`
}

const kiwiAttrKey = "key"
const kiwiAttrPattern = "pattern"
