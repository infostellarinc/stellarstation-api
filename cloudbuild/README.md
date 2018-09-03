This directory contains configuration for google cloud build. Currently this is only used to
publish snapshots for the Java and golang API stubs.

To re-generate the encrypted github API key:

```bash
gcloud kms encrypt --plaintext-file=id_rsa --ciphertext-file=id_rsa.enc --location=us-central1 --keyring=cloudbuild --key=deploy-stubs
```

The public portion of the key should be added to the 'deploy keys' section of the go-stellarstation repo.

To re-generate the bintray API key:

```bash
echo -n KEY | gcloud kms encrypt --plaintext-file=- --ciphertext-file=- --location=us-central1 --keyring=cloudbuild --key=deploy-stubs | base64 -w 0
```

Changes to the config can be tested locally with:

```bash
cloud-build-local -config cloudbuild/snapshot-publish.yaml --dryrun=false .
```
