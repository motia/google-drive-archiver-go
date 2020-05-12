A project that downloads Google Drive contents and uploads them zipped to S3 

# Google Drive
1- Visit https://developers.google.com/drive/api/v3/quickstart/nodejs, and create an API key, then download `credentials.json` which contains your Google Drive API key.
2- Activate sharing for the Drive directory you want to backup
3- Add config keys of Google Drive `config.json` in the root of the project
```
  "googleDriveRoot": id of the root Drive directory,
  "googleDriveCredentials": content of 'credentials.json'
  }
```

# S3
1- Add S3 entries to config.json
```
  "s3": {
    "accessKey": "ACCESS_KEY",
    "secretKey": "SECRET_KEY",
    "bucketName": "BUCKET_NAME"
  },
```
