cf set-env watchman CF_DOMAIN 10.244.0.34.xip.io
cf set-env watchman STATSD_ADDRESS 10.244.2.2:8125
cf set-env watchman STATSD_PREFIX CF-
cf set-env watchman FIREHOSE_SUBSCRIPTION_ID watchman
cf set-env watchman CLIENT_SECRET watchman
cf set-env watchman CLIENT_ID watchman
