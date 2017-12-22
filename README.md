Not applicable

#### Authentication - service account (recommended)

On the console create a service acount for Kuberentes Engine.

* Go to the API Console Credentials page.
* From the project drop-down, select your project.
* On the Credentials page, select the Create credentials drop-down, then select Service account key.
* From the Service account drop-down, select an existing service account or create a new one.
* For Key type, select the JSON key option, then select Create. The file automatically downloads to your computer.
* Put the *.json file you just downloaded in a directory of your choosing. 
* Set and export the environment variable GOOGLE_APPLICATION_CREDENTIALS to the path of the JSON file downloaded.

For further info follow this [link](https://developers.google.com/identity/protocols/application-default-credentials)

#### Authentication - gcloud

Easiest auth during development is to use the `gcloud` cli tool. To install it:

```
brew tap caskroom/cask
brew cask install google-cloud-sdk
```

Once gcloud is installed the easiest way is to authenticate from where running the binary: `gcloud auth application-default login`.

#### Clusters

You can `create`, `delete` and `update` clusters.


