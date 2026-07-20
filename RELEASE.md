# Creating a release

- Set the version in frontend/package.json
- Create a github release
- Publish the docker images (see below)

## Releasing the docker images

- Build the docker images with `task docker-build-api` and `task docker-build-api`

- Tag the images with the version number

```
docker image tag ekse/rossoreader-api:latest ekse/rossoreader-api:0.2
docker image tag ekse/rossoreader-web:latest ekse/rossoreader-web:0.2
```

- Publish the images (version and latest)

```
docker push ekse/rossoreader-web:0.2
docker push ekse/rossoreader-web:latest
docker push ekse/rossoreader-api:0.2
docker push ekse/rossoreader-api:latest
```