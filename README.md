#### Authentication

Easiest auth during development is to use the `gcloud` cli tool. To install it:

```
brew tap caskroom/cask
brew cask install google-cloud-sdk
```

Once gcloud is installed the easiest way is to authenticate from where running the binary: `gcloud auth application-default login`.

There are other ways to initiate credentials and auth: 
Initialize credentials - https://developers.google.com/identity/protocols/application-default-credentials

#### Clusters

You can `create`, `delete` and `update` clusters with: 

```
./gke-test -project $PROJECT_ID -zone us-central1-a
```


