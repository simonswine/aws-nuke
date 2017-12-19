package resources

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
)

type S3ObjectVersion struct {
	svc       *s3.S3
	bucket    string
	versionId string
	key       string
}

func (e *S3ObjectVersion) Remove() error {
	params := &s3.DeleteObjectInput{
		VersionId: &e.versionId,
		Bucket:    &e.bucket,
		Key:       &e.key,
	}

	_, err := e.svc.DeleteObject(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *S3ObjectVersion) String() string {
	return fmt.Sprintf("s3://%s/%s#%s", e.bucket, e.key, e.versionId)
}
