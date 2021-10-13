[![ContainerSSH - Launch Containers on Demand](https://containerssh.github.io/images/logo-for-embedding.svg)](https://containerssh.io/)

<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">ContainerSSH Test Helper Library</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/containerssh/test?style=for-the-badge)](https://goreportcard.com/report/github.com/containerssh/test)
[![LGTM Alerts](https://img.shields.io/lgtm/alerts/github/ContainerSSH/test?style=for-the-badge)](https://lgtm.com/projects/g/ContainerSSH/test/)

This library helps with bringing up services for testing, such as S3, oAuth, etc. **All services require an exposed Docker socket to work.**

<p align="center"><strong>⚠⚠⚠ Warning: This is a developer documentation. ⚠⚠⚠</strong><br />The user documentation for ContainerSSH is located at <a href="https://containerssh.io">containerssh.io</a>.</p>

## Starting an S3 service

To start the S3 service and then use it with the AWS client as follows:

```go
package your_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/containerssh/test"
)

func TestYourFunc(t *testing.T) {
	s3Service := test.S3(t)

	awsConfig := &aws.Config{
		Credentials: credentials.NewCredentials(
			&credentials.StaticProvider{
				Value: credentials.Value{
					AccessKeyID:     s3Service.AccessKey(),
					SecretAccessKey: s3Service.SecretKey(),
				},
			},
		),
		Endpoint:         aws.String(s3Service.URL()),
		Region:           aws.String(s3Service.Region()),
		S3ForcePathStyle: aws.Bool(s3Service.PathStyle()),
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		t.Fatalf("failed to establish S3 session (%v)", err)
	}
	s3Connection := s3.New(sess)
	
	// ...
}
```

That's it! Now you have a working S3 connection for testing!
