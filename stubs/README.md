# API stubs

This directory builds language specific stubs from the proto definition. This documentation is mainly
intended to be read by Infostellar employees.

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

We only support 'release' versions of the stubs in Java, but 'snapshot' versions are available
for development and testing between releases. Please use snapshots at your own risk.

### Release

TODO: Needs update for new publishing pipeline.

#### maven central

The stubs are hosted on maven central.

### Snapshot

Snapshot builds are hosted on Sonatype's OSS repo:

- https://oss.sonatype.org/#nexus-search;quick~stellarstation-api

This repository is added to the gradle list of repos by the curiostack plugin. If you don't use that
plugin you'll need to add a maven repository pointing at 'https://oss.sonatype.org/content/repositories/snapshots'
to your gradle build.

The following command is used to do the publishing:

```
./gradlew :api:publishMavenPublicationToMavenRepository
  -Pmaven.username=<sonatype_ossrh_username>
  -Pmaven.password="<sonatype_ossrh_password>"
  -Psigning.secretKeyRingFile=<gpg_keyring_absolute_file_path>
  -Psigning.keyId=C8772BE9
  -Psigning.password="<gpg_key_password>"
```

## C++
## golang
## nodejs

