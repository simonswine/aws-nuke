package aws_nuke

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/simonswine/aws-nuke/pkg/aws_nuke/interfaces"
	"github.com/simonswine/aws-nuke/pkg/aws_nuke/s3"
)

const FlagForceDestroy = "force-destroy"

var _ interfaces.AWSNuke = &AWSNuke{}

type AWSNuke struct {
	log      *logrus.Entry
	sessions map[interfaces.Region]*session.Session
}

func New() *AWSNuke {
	return &AWSNuke{
		sessions: make(map[interfaces.Region]*session.Session),
	}
}

func (a *AWSNuke) Session(region interfaces.Region) *session.Session {
	if sess, ok := a.sessions[region]; ok {
		return sess
	}

	a.sessions[region] = session.Must(session.NewSession(&aws.Config{Region: aws.String(string(region))}))

	return a.sessions[region]
}

func (a *AWSNuke) Log() *logrus.Entry {
	if a.log == nil {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel
		a.log = logger.WithField("app", "aws-nuke")
	}
	return a.log
}

func (a *AWSNuke) Must(err error) {
	if err != nil {
		a.log.Fatal(err)
	}
}

func (a *AWSNuke) CmdS3(cmd *cobra.Command, args []string) error {
	myS3 := s3.New(a)

	destroy, err := cmd.Flags().GetBool(FlagForceDestroy)
	if err != nil {
		a.log.Warn("unexpected error getting flag: %s", err)
		destroy = false
	}

	if destroy {
		if err := myS3.Delete(); err != nil {
			return err
		}
	} else {
		list, err := myS3.List()
		if err != nil {
			return err
		}
		msg := "s3 buckets to delete: %s"
		a.log.Infof(msg, list)
	}

	return nil
}
