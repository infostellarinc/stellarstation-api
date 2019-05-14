# API stubs

This directory builds language specific stubs from the proto definition.

## Versioning

The version number applied to the stubs is determined by the curiostack gradle plugin. For
builds marked as 'release' the version number comes from one of the following (checked in
order):

- REVISION_ID env var
- TAG_NAME env var
- BRANCH_NAME env var
- most recent git commit ID

For 'non-release' (i.e. snapshot)  builds, the version is set to the form
0.0.0-TIMESTAMP-id (where id is determined as per release builds above).

## Java

We support 'release' and 'snapshot' versions of the stubs in Java.

### Release

```
./gradlew -Pcuriostack.release=true :api:bintrayUpload
```

Normally this is not invoked directly but is instead run automatically from google cloud build (which
deals with setting up the environment etc).

Note that bintray.user and bintray.key properties must be set (istellar users can find the
details in the usual place). Environment variables should probably also be set to ensure
an appropriate version number is generated.

#### jcenter

The published stubs are hosted by bintray and available through jcenter automatically:
- https://jcenter.bintray.com/com/stellarstation/api/stellarstation-api/

Infostellar employees can log in to bintray to manage the packages which have already been released etc.

#### maven central

The stubs are also hosted on maven central, but a manual step is required here. The infostellar
sysadmin must log in to bintray (account details listed in the usual place) and manually select
'maven central' for the most recently updated release (this should cause it to be synced to maven
central).

### Snapshot

Snapshot builds are hosted on an artifactory repo:

- https://oss.jfrog.org/libs-snapshot/com/stellarstation/api/stellarstation-api/

This repository is added to the gradle list of repos by the curiostack plugin. If you don't use that
plugin you'll need to add a maven repository pointing at 'https://oss.jfrog.org/artifactory/oss-snapshot-local'
to your gradle build (using oss.jfrog.org/libs-snapshot as the URL may also work).

The following command is used to do the publishing:

```
 ./gradlew :api:artifactoryPublish
```

The above is run automatically by google cloud build whenever a PR is merged to master. As with the
release build, bintray.user and bintray.key properties must be set.

## C++
## golang
## nodejs

