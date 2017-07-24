package s3

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"

	"github.com/simonswine/aws-nuke/pkg/aws_nuke/interfaces"
)

type Buckets struct {
	a            interfaces.AWSNuke
	log          *logrus.Entry
	services     map[interfaces.Region]*s3.S3
	servicesLock sync.Mutex
	buckets      []*bucket
}

type bucket struct {
	*s3.Bucket
	region *interfaces.Region
	b      *Buckets
	keep   *bool
}

func (b *bucket) Delete() error {
	keep, err := b.Keep()
	if err != nil {
		return err
	}

	if keep {
		b.b.log.WithField("bucket", *b.Name).Debug("ignoring bucket")
		return nil
	}

	region, err := b.Region()
	if err != nil {
		return err
	}

	svc := b.b.service(region)
	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: b.Name,
	})

	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok && ec2err.Code() == "BucketNotEmpty" {
			resp, err := svc.ListObjectVersions(
				&s3.ListObjectVersionsInput{
					Bucket: b.Name,
				},
			)

			if err != nil {
				return fmt.Errorf("error listing objects to delete: %s", err)
			}

			objectsToDelete := make([]*s3.ObjectIdentifier, 0)

			if len(resp.DeleteMarkers) != 0 {

				for _, v := range resp.DeleteMarkers {
					objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			if len(resp.Versions) != 0 {
				for _, v := range resp.Versions {
					objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			params := &s3.DeleteObjectsInput{
				Bucket: b.Name,
				Delete: &s3.Delete{
					Objects: objectsToDelete,
				},
			}

			_, err = svc.DeleteObjects(params)

			if err != nil {
				return fmt.Errorf("error deleting objects: %s", err)
			}

			// recurse until bucket empty & deleted
			return b.Delete()
		} else {
			return err
		}
	}

	return nil

	/*

		log.Printf("[DEBUG] S3 Delete Bucket: %s", d.Id())
		_, err := s3conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(d.Id()),
		})
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if ok && ec2err.Code() == "BucketNotEmpty" {
				if d.Get("force_destroy").(bool) {
					// bucket may have things delete them
					log.Printf("[DEBUG] S3 Bucket attempting to forceDestroy %+v", err)

					bucket := d.Get("bucket").(string)

					// this line recurses until all objects are deleted or an error is returned
					return resourceAwsS3BucketDelete(d, meta)
				}
			}
			return fmt.Errorf("Error deleting S3 Bucket: %s %q", err, d.Get("bucket").(string))

	*/
}

func (b *bucket) Region() (interfaces.Region, error) {
	if b.region == nil {
		location, err := b.b.service(interfaces.DefaultRegion).GetBucketLocation(&s3.GetBucketLocationInput{Bucket: b.Name})
		if err != nil {
			return "", err
		}

		region := interfaces.Region("us-east-1")
		if location.LocationConstraint != nil {
			region = interfaces.Region(*location.LocationConstraint)
		}
		b.region = &region
	}

	return *b.region, nil
}

func (b *bucket) Keep() (bool, error) {
	valTrue := true
	valFalse := false

	if b.keep == nil {
		region, err := b.Region()
		if err != nil {
			return false, err
		}

		tags, err := b.b.service(region).GetBucketTagging(&s3.GetBucketTaggingInput{
			Bucket: b.Name,
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "NoSuchTagSet" {
					b.keep = &valFalse
					return *b.keep, nil
				}
			}

			return true, err
		}

		for _, t := range tags.TagSet {
			if t.Key != nil && *t.Key == interfaces.TagKeepKey && t.Value != nil && *t.Value == interfaces.TagKeepValue {
				b.keep = &valTrue
				return *b.keep, nil
			}
		}

		b.keep = &valFalse
		return *b.keep, nil
	}

	return *b.keep, nil
}

func New(a interfaces.AWSNuke) *Buckets {
	return &Buckets{
		a:        a,
		log:      a.Log().WithField("module", "s3"),
		services: map[interfaces.Region]*s3.S3{},
	}
}

func (b *Buckets) service(region interfaces.Region) *s3.S3 {
	b.servicesLock.Lock()
	if _, ok := b.services[region]; !ok {
		b.services[region] = s3.New(b.a.Session(region))
	}
	svc, _ := b.services[region]
	b.servicesLock.Unlock()
	return svc
}

func (b *Buckets) Buckets() ([]*bucket, error) {
	if len(b.buckets) == 0 {
		buckets, err := b.service(interfaces.DefaultRegion).ListBuckets(&s3.ListBucketsInput{})
		if err != nil {
			return []*bucket{}, err
		}

		b.buckets = make([]*bucket, len(buckets.Buckets))
		for pos, _ := range buckets.Buckets {
			b.buckets[pos] = &bucket{
				Bucket: buckets.Buckets[pos],
				b:      b,
			}
		}
	}
	return b.buckets, nil
}

func (b *Buckets) List() ([]string, error) {
	buckets, err := b.Buckets()
	if err != nil {
		return []string{}, err
	}

	var bucketNames []string
	var bucketNamesLock sync.Mutex
	var wg sync.WaitGroup

	for pos, _ := range buckets {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			k, err := buckets[pos].Keep()
			if err != nil {
				b.log.Warnf("error getting S3 tags: %s", err)
				return
			}
			if !k {
				bucketNamesLock.Lock()
				bucketNames = append(bucketNames, *buckets[pos].Name)
				bucketNamesLock.Unlock()
			}
		}(pos)
	}
	wg.Wait()
	return bucketNames, nil
}

func (b *Buckets) Delete() error {
	buckets, err := b.Buckets()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for pos, _ := range buckets {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			err := buckets[pos].Delete()
			if err != nil {
				b.log.WithField("bucket", *buckets[pos].Name).Warnf("error deleting bucket: %s", err)
				return
			}
		}(pos)
	}
	wg.Wait()
	return nil
}
