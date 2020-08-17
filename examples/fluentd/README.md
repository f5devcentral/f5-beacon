# Fluentd integration with Beacon

[Fluentd](https://www.fluentd.org/) is a great tool to unify data collection and consumption.  There are a significant number of existing [plugins](https://www.fluentd.org/plugins/all) to integrate various sources.  In addition, there are plugins for filtering and parsing data within Fluentd before routing to an output.  The [Beacon plugin](https://github.com/f5devcentral/fluent-plugin-f5-beacon) feeds data to Beacon via the same mechanism as [Telegraf](https://www.influxdata.com/time-series-platform/telegraf/).  This means the Fluentd data is mapped to measurements, tags and fields for ingestion into Beacon.

The measurement is determined either via the via the `measurement` parameter within the Fluentd configuration for the Beacon plugin, or via the Fluentd tag.  The match directive for the Beacon plugin defines what Fluentd tags are routed to the Beacon output plugin.  Data within Fluentd events are mapped to tags and fields in a similar fashion.  If the `auto_tags` parameter is used for the Beacon plugin, then string values are mapped to tags and all else are mapped to fields.  The alternative is to use the `tag_keys` parameter to define the data elements that are mapped to tags.

If the Fluentd data contains more complex data elements (arrays, nested objects), the data will typically need to be passed through additional Fluentd plugins to make it usable by Beacon.  By default, the Beacon plugin discards such elements.  When this happens, it will be noted in the Fluentd log:

```
2020-08-10 22:11:16 +0000 [warn]: #0 [f5_beacon] array/hash field 'data' discarded; consider using a plugin to map
```

## Splitting arrays

Imagine a Fluentd event that looks similar to the below:

```
{
    "numRecords": 3,
    "metadata": {
        "streamId": "1116"
    },
    "data": [
        {
            "1xx": 0,
            "2xx": 35,
            "3xx": 0,
            "4xx": 5,
            "5xx": 0
        },
        {
            "1xx": 0,
            "2xx": 41,
            "3xx": 0,
            "4xx": 2,
            "5xx": 0
        },
        {
            "1xx": 0,
            "2xx": 38,
            "3xx": 0,
            "4xx": 0,
            "5xx": 0
        }
    ]
}
```

The key metric data is embedded within the `data` array, which contains three separate records.  In Beacon, we'd like to generate insights based on the HTTP status distribution in these records.  Instead of routing the data to the Beacon plugin directly, we can first route it to another Fluentd plugin, to split the event (one event per array item).  Using the [fluent-plugin-record_splitter](https://github.com/ixixi/fluent-plugin-record_splitter) plugin as an example, the configuration might look similar to:

```
<match pattern>
  @type record_splitter
  tag foo.split
  split_key data
  keep_other_key false
</match>
```

The Beacon plugin can then match `foo.split` to receive the individual records for processing.

There are a handful of output/filter plugins that do similar manipulations and can be used for various use cases.

## Flattening nested objects

Imagine the prior example with a slight modification where each record has an additional level of nesting that details the distribution at each level:

```
{
    "numRecords": 3,
    "metadata": {
        "streamId": "1116"
    },
    "data": [
        {
            "1xx": 0,
            "2xx": 35,
            "3xx": 0,
            "4xx": 5,
            "5xx": 0,
            "2xx_dist": {
                "200": 35
            },
            "4xx_dist": {
                "401": 2,
                "404": 3
            }
        },
        ...
    ]
}
```

To retain the distributions for consumption within Beacon, we can flatten the distribution hashes to become top-level fields within each record.  For this case, we use the [fluent-plugin-flatten-hash](https://github.com/kazegusuri/fluent-plugin-flatten-hash) plugin as an example.  The configuration might look similar to:

```
<match pattern>
  @type flatten_hash
  add_tag_prefix flat.
  separator _
</match>
```

In this example, the plugin adds a `flat.` prefix onto the tag, which the Beacon plugin can then match.

The resulting data sent to Beacon using both the splitting/flattening plugins noted above would look similar to:

```
{
    "1xx": 0,
    "2xx": 35,
    "3xx": 0,
    "4xx": 5,
    "5xx": 0,
    "2xx_dist_200": 35,
    "4xx_dist_401": 2,
    "4xx_dist_404": 3
},
```

## Specifying timestamp

By default, the timestamp sent to Beacon will be Fluentd's event timestamp.  If the data has a timestamp field, it can be used instead.  There are a couple of ways this can happen.  The recommended approach if possible is to parse the time on input into the Fluentd's event timestamp.  See [Config: Parse Section](https://docs.fluentd.org/configuration/parse-section) for more details.  It may be that parsing the time on input isn't possible because the data requires additional transformation first.  For example, imagine a timestamp field on each record in the example above:

```
{
    "numRecords": 3,
    "metadata": {
        "streamId": "1116"
    },
    "data": [
        {
            "1xx": 0,
            "2xx": 35,
            "3xx": 0,
            "4xx": 5,
            "5xx": 0,
            "2xx_dist": {
                "200": 35
            },
            "4xx_dist": {
                "401": 2,
                "404": 3
            },
            "recordTime": "2020-05-29T17:02:00Z"
        },
        ...
    ]
}
```

In this case, we'd like to still apply the splitting and flattening discussed previously and afterwards, use `recordTime` as the event timestamp.  To address this, we could apply a filter that parses the time before processing by the Beacon plugin.  For this case, we use the [fluent-plugin-filter_typecast](https://github.com/sonots/fluent-plugin-filter_typecast) filter plugin as an example.  The configuration might look similar to:

```
<filter pattern>
  @type typecast
  types recordTime:time:%Y-%m-%dT%H:%M:%S%Z
</filter>
```

The `pattern` used would be the same used by the Beacon plugin.  As a final step, we provide an additional parameter to the Beacon plugin so that it uses the parsed timestamp instead of the event timestamp.

```
time_key recordTime
```