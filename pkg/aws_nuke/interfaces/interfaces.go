package interfaces

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
)

const DefaultRegion = Region("eu-west-1")

const TagKeepKey = "awsnuke_keep"
const TagKeepValue = "true"

type Region string

type AWSNuke interface {
	Session(region Region) *session.Session
	Log() *logrus.Entry
}
