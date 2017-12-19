package resources

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Object struct {
	svc    *s3.S3
	bucket string
	key    string
}

func (n *S3Nuke) ListObjects() ([]Resource, error) {
	resources := make([]Resource, 0)

	buckets, err := n.DescribeBuckets()
	if err != nil {
		return nil, err
	}

	for _, name := range buckets {
		params := &s3.ListObjectVersionsInput{
			Bucket: &name,
		}

		for {
			resp, err := n.Service.ListObjectVersions(params)
			if err != nil {
				return nil, err
			}

			for _, out := range resp.Versions {
				if out.VersionId != nil && *out.VersionId != "null" {
					resources = append(resources, &S3ObjectVersion{
						svc:       n.Service,
						bucket:    name,
						key:       *out.Key,
						versionId: *out.VersionId,
					})
				} else {
					resources = append(resources, &S3Object{
						svc:    n.Service,
						bucket: name,
						key:    *out.Key,
					})
				}
			}

			for _, out := range resp.DeleteMarkers {
				if out.VersionId != nil && *out.VersionId != "null" {
					resources = append(resources, &S3ObjectVersion{
						svc:       n.Service,
						bucket:    name,
						key:       *out.Key,
						versionId: *out.VersionId,
					})
				} else {
					resources = append(resources, &S3Object{
						svc:    n.Service,
						bucket: name,
						key:    *out.Key,
					})
				}
			}

			// make sure to list all with more than 1000 objects
			if *resp.IsTruncated {
				params.KeyMarker = resp.NextKeyMarker
				continue
			}

			break
		}
	}

	return resources, nil
}

func (e *S3Object) Remove() error {
	params := &s3.DeleteObjectInput{
		Bucket: &e.bucket,
		Key:    &e.key,
	}

	_, err := e.svc.DeleteObject(params)
	if err != nil {
		return err
	}

	return nil
}

func (e *S3Object) String() string {
	return fmt.Sprintf("s3://%s/%s", e.bucket, e.key)
}
