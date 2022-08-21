package config

type Bucket struct {
	RomBucket   string
	ImageBucket string
	VideoBucket string
}

type Config struct {
	BucketInfo *Bucket
}
