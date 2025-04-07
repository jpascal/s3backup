### Configure environment
```shell
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_DEFAULT_REGION
AWS_ENDPOINT_URL
```

### List backups

```shell
s3backup --bucket=<name> list 
```

### Create new backup

```shell
s3backup --bucket=<name> create --src=<pattern> --dst=<directory> --group=week --clean --keep=2
```

### Delete backups

```shell
s3backup --bucket=<name> delete --prefix=<directory|object>
```

### Clean backups

```shell
s3backup --bucket=<name> clean --dst=<directory> --group=week --keep=0
```

### Build
```shell
docker build . --platform linux/amd64 -t jpascal/s3backup:v0.0.6
docker push jpascal/s3backup:v0.0.6
```


### Commands
```
Flags:
  -h, --help             Show context-sensitive help.
      --bucket=STRING    bucket name

Commands:
  create [flags]
    do backup

  list [flags]
    list files

  clean [flags]
    clean backups

  delete [flags]
    delete files by prefix

```

##### Command: create
```
Flags:
  -h, --help             Show context-sensitive help.
      --bucket=STRING    bucket name

      --src="*"          source files
      --group=STRING     group name
      --partSize=1000    part size in Mb
      --dst="/"
      --clean            clean old backups using keep value
      --keep=UINT        keep number of files
```

#### Command: list 
```
Flags:
  -h, --help             Show context-sensitive help.
      --bucket=STRING    bucket name

      --prefix=          prefix
```

#### Command: clean
```
Flags:
  -h, --help             Show context-sensitive help.
      --bucket=STRING    bucket name

      --group=STRING     group name
      --keep=UINT        keep number of files
```

#### Command: delete
```
Flags:
  -h, --help             Show context-sensitive help.
      --bucket=STRING    bucket name

      --prefix=          prefix
```
