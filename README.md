Watchman
========

A firehose nozzle that runs in Cloud Foundry for monitoring HTTP endpoints. Metrics for each endpoint are dumped to StatsD. Nozzle's are designed to scale horizontally, have no state and are thus a great fit for running inside of the platform. Nozzles can scale to distribute load, CF knows how to do that, seems like a win win. This repo is an experiment in that vein, what needs to be done to run a nozzle in CF, what are the consequences of doing so?

A ton of thanks to [CloudCredo](http://cloudcredo.com/) for having done the [heavy lifting](http://cloudcredo.com/how-to-integrate-graphite-with-cloud-foundry/) of parsing the firehose, and then writing about. This is really just an adaptation of [Ed King's work](https://github.com/CloudCredo/graphite-nozzle).

[![wercker status](https://app.wercker.com/status/a92c776ac16c25caaff0027933e352e1/s "wercker status")](https://app.wercker.com/project/bykey/a92c776ac16c25caaff0027933e352e1)

#Running

First, you need a StatsD instance to pump to and something to visualize it with. I've been using the [bosh release for StatsD + Graphite from CloudCredo](https://github.com/CloudCredo/graphite-statsd-boshrelease). It's fantastic and a great place to start.

This nozzle relies on a few environment variables: 

* `CLIENT_ID` Is the uaa client we should authenticate as.
* `CLIENT_SECRET` Is the secret for uaa client so we can obtain a token.
* `UAA_ADDRESS` Is the uaa instance we should talk to for tokens.
* `DOPPLER_ADDRESS` Is where the nozzle attaches to to receive metrics
* `STATSD_ADDRESS` Is where the nozzle sends metrics to 
* `STATSD_PREFIX` Controls the StatsD node names that various metrics will appear under. 
* `FIREHOSE_SUBSCRIPTION_ID` Tell the firehose who's connecting

## Create a user in uaa to watch the firehose
You shouldn't really use admin tokens to watch the firehose, so lets create something with limited access. 

```
$ uaac client add watchman --scope uaa.none --authorized_grant_types "client_credentials" --authorities doppler.firehose --redirect_uri http://example.com 

# This will return a bearer token.
$ curl -k -v 'https://watchman:watchman@uaa.10.244.0.34.xip.io/oauth/token?grant_type=client_credentials'
```

 
## Deploying

Push, set some environment variables, start it up. Make sure you're quick about the oauth token bit as they expire!

```
cf push --no-start
cf set-env watchman DOPPLER_ADDRESS wss://doppler.10.244.0.34.xip.io:443
cf set-env watchman STATSD_ADDRESS 10.244.2.2:8125
cf set-env watchman STATSD_PREFIX CloudFoundry
cf set-env watchman FIREHOSE_SUBSCRIPTION_ID WatchmanFirehose
cf set-env watchman CLIENT_SECRET watchman
cf set-env watchman CLIENT_ID watchman
cf set-env watchman UAA_ADDRESS https://uaa.10.244.0.34.xip.io
cf start watchman
```

A word of warning, if you're using bosh-lite make sure you change the default security groups to allow routing to the 10.0.0.0/16 network. 

##TODO

* Statsd, doppler, and uaa should be exposed as services (`cf cups`).



