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

### Build controller and proxy
```shell
docker build . --platform linux/amd64 -t jpascal/s3backup:v0.0.3
docker push jpascal/s3backup:v0.0.3
