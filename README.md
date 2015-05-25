Watchman
========

A firehose nozzle that runs in CloudFoundry for monitoring HTTP endpoints. Metrics for each endpoint are dumped to StatsD.

#Running

This nozzle relies on a few environment variables: 


* `CF_ACCESS_TOKEN` Allows the nozzle to register with CF and receive events
* `DOPPLER_ADDRESS` Is where the nozzle attaches to to receive metrics
* `STATSD_ADDRESS` Is where the nozzle sends metrics to 
* `STATSD_PREFIX` Controls the StatsD node that various metrics will appear under. 

## Create a user in uaa to watch the firehose
You shouldn't really use admin tokens to watch the firehose, so lets create something with limited access. 

```
$ uaac client add watchman --scope uaa.none --authorized_grant_types "authorization_code, client_credentials, refresh_token" --authorities doppler.firehose --redirect_uri http://example.com 
``` 
## Deploying

Push, set some environment variables, start it up. Make sure you're quick about the oauth token bit as they expire!

```
cf push --no-stat
cf set-env watchman DOPPLER_ADDRESS wss://doppler.10.244.0.34.xip.io:443
cf set-env watchman STATSD_ADDRESS 10.244.2.2:8125
cf set-env watchman STATSD_PREFIX CloudFoundry
cf set-env watchman FIREHOSE_SUBSCRIPTION_ID WatchmanFirehose
cf set-env watchman CF_ACCESS_TOKEN "`uaac context watchman | grep access_token | sed -e "s/access_token:/bearer/"`"
```

##TODO

* Refresh token
* Subscription ID + Distributed load via app index