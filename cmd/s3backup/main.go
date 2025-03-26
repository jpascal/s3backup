package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rodaine/table"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"slices"
)

func prettyByteSize(b int64) string {
	ss := float64(b)
	for _, unit := range []string{"", "K", "M", "G", "T", "P", "E", "Z"} {
		if math.Abs(ss) < 1024.0 {
			return fmt.Sprintf("%3.1f%sb", ss, unit)
		}
		ss /= 1024.0
	}
	return fmt.Sprintf("%.1fYb", ss)
}

type CLI struct {
	Bucket string        `help:"bucket name" name:"bucket"`
	Create CreateCommand `cmd:"create" help:"do backup"`
	List   ListCommand   `cmd:"list" help:"list files"`
	Clean  CleanCommand  `cmd:"clean" help:"clean backups"`
	Delete DeleteCommand `cmd:"delete" help:"delete files by prefix"`
}

type CleanCommand struct {
	Group       string `help:"group name" name:"group"`
	Keep        uint   `help:"keep number of files" name:"keep"`
	Destination string `name:"dst" default:"/"`
}

func (l *CleanCommand) Run(ctx *Context) error {
	output, err := ctx.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(ctx.Bucket),
		Prefix: aws.String(l.Destination),
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var collected []*s3.Object

	for _, item := range output.Contents {
		head, err := ctx.S3.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(ctx.Bucket),
			Key:    item.Key,
		})
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		if value, ok := head.Metadata["group"]; ok && *value == l.Group {
			collected = append(collected, item)
		}
	}

	slices.SortFunc(collected, func(a, b *s3.Object) int {
		at, bt := *a.LastModified, *b.LastModified
		return at.Compare(bt)
	})
	if len(collected) > int(l.Keep) {
		for _, item := range collected[l.Keep:] {
			slog.Info(fmt.Sprintf("delete s3://%s (size: %s)", filepath.Join(ctx.Bucket, *item.Key), prettyByteSize(*item.Size)))
			if _, err := ctx.S3.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(ctx.Bucket),
				Key:    item.Key,
			}); err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		}
	}
	return nil
}

type CreateCommand struct {
	Source      string `name:"src" help:"source files" default:"*"`
	Group       string `help:"group name" name:"group"`
	Destination string `name:"dst" default:"/"`
	Clean       bool   `name:"clean" help:"clean old backups using keep value" default:"false"`
	Keep        uint   `help:"keep number of files" name:"keep"`
}

func (l *CreateCommand) Run(ctx *Context) error {
	filePaths, err := filepath.Glob(l.Source)
	if err != nil {
		return err
	}
	for _, filePath := range filePaths {
		fileStats, err := os.Lstat(filePath)
		if err != nil {
			return err
		}
		if fileStats.IsDir() {
			continue
		}
		slog.Info(fmt.Sprintf("upload %s to s3://%s (size: %s, group: %s)", filePath, filepath.Join(ctx.Bucket, l.Destination, filePath), prettyByteSize(fileStats.Size()), l.Group))
		if file, err := os.Open(filePath); err != nil {
			return err
		} else if _, err := ctx.S3.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(ctx.Bucket),
			Key:    aws.String(filepath.Join(l.Destination, filePath)),
			Body:   file,
			Metadata: map[string]*string{
				"group": aws.String(l.Group),
			},
		}); err != nil {
			return err
		}
	}
	if l.Clean {
		cleanCommand := &CleanCommand{
			Group:       l.Group,
			Keep:        l.Keep,
			Destination: l.Destination,
		}
		return cleanCommand.Run(ctx)
	}
	return nil
}

type Context struct {
	Bucket string
	Group  string
	S3     *s3.S3
}

type ListCommand struct {
	Prefix *string `help:"prefix" default:""`
}

func (l *ListCommand) Run(ctx *Context) error {
	output, err := ctx.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(ctx.Bucket),
		Prefix: l.Prefix,
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	tbl := table.New("Name", "Last modified", "Size", "Group")
	for _, item := range output.Contents {
		head, err := ctx.S3.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(ctx.Bucket),
			Key:    item.Key,
		})
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		group := "-"
		if value, ok := head.Metadata["group"]; ok {
			group = *value
		}
		tbl.AddRow(*item.Key, *item.LastModified, prettyByteSize(*item.Size), group)
	}
	tbl.Print()
	return nil
}

type DeleteCommand struct {
	Prefix *string `help:"prefix" default:""`
}

func (l *DeleteCommand) Run(ctx *Context) error {
	output, err := ctx.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(ctx.Bucket),
		Prefix: l.Prefix,
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	for _, item := range output.Contents {
		slog.Info(fmt.Sprintf("delete s3://%s (size: %s)", filepath.Join(ctx.Bucket, *item.Key), prettyByteSize(*item.Size)))
		if _, err := ctx.S3.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(ctx.Bucket),
			Key:    item.Key,
		}); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	return nil
}

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(os.Getenv("AWS_DEFAULT_REGION")),
		Endpoint:    aws.String(os.Getenv("AWS_ENDPOINT_URL")),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	sess.Config.LowerCaseHeaderMaps = aws.Bool(true)
	cli := &CLI{}
	ctx := kong.Parse(cli)
	err = ctx.Run(&Context{Bucket: cli.Bucket, S3: s3.New(sess)})
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
