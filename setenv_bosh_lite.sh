cf set-env watchman DOPPLER_ENDPOINT wss://doppler.10.244.0.34.xip.io:443
cf set-env watchman UAA_ENDPOINT https://uaa.10.244.0.34.xip.io
cf set-env watchman STATSD_ADDRESS 10.244.2.2:8125
cf set-env watchman STATSD_PREFIX CF-
cf set-env watchman FIREHOSE_SUBSCRIPTION_ID watchman
cf set-env watchman CLIENT_SECRET watchman
cf set-env watchman CLIENT_ID watchman
