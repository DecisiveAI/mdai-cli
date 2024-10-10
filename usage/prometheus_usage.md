# Prometheus queries

Can be queried directly from prometheus by hitting the endpoint `http://localhost:{PROMETHEUS_PORT}/api/v1/query?query={QUERY}` for instant queries or `http://localhost:{PROMETHEUS_PORT}/api/v1/query_range?query={QUERY}` for range queries. Alternatively, you can use the Prometheus [Expression browser](https://prometheus.io/docs/visualization/browser/) to see a visualization of the data that these queries return.

See the [Prometheus docs](https://prometheus.io/docs/prometheus/latest/querying/basics/) for information on how to modify queries.
4

## Analysis tab


All in one query [LINK TO PROMETHEUS](http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(service_name%2C%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(logs%7Cspans%7Cmetrics)_sent_total%22%7D%2C%0A%20%20%20%20%20%20%22data_type%22%2C%0A%20%20%20%20%20%20%22%241%22%2C%0A%20%20%20%20%20%20%22__name__%22%2C%0A%20%20%20%20%20%20%22mdai_(logs%7Cspans%7Cmetrics)_sent_total%22%0A%20%20%20%20)%0A%20%20)%5B1d%3A%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h&g1.expr=increase(%0A%20%20sum%20by%20(service_name%2C%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(logs%7Cspans%7Cmetrics)_received_total%22%7D%2C%0A%20%20%20%20%20%20%22data_type%22%2C%0A%20%20%20%20%20%20%22%241%22%2C%0A%20%20%20%20%20%20%22__name__%22%2C%0A%20%20%20%20%20%20%22mdai_(logs%7Cspans%7Cmetrics)_received_total%22%0A%20%20%20%20)%0A%20%20)%5B1d%3A%5D%0A)&g1.tab=0&g1.display_mode=lines&g1.show_exemplars=0&g1.range_input=1h&g2.expr=increase(%0A%20%20sum%20by%20(service_name%2C%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_received_total%22%7D%2C%0A%20%20%20%20%20%20%22data_type%22%2C%0A%20%20%20%20%20%20%22%241%22%2C%0A%20%20%20%20%20%20%22__name__%22%2C%0A%20%20%20%20%20%20%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_received_total%22%0A%20%20%20%20)%0A%20%20)%5B1d%3A%5D%0A)&g2.tab=0&g2.display_mode=lines&g2.show_exemplars=0&g2.range_input=1h&g3.expr=increase(%0A%20%20sum%20by%20(service_name%2C%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_sent_total%22%7D%2C%0A%20%20%20%20%20%20%22data_type%22%2C%0A%20%20%20%20%20%20%22%241%22%2C%0A%20%20%20%20%20%20%22__name__%22%2C%0A%20%20%20%20%20%20%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_sent_total%22%0A%20%20%20%20)%0A%20%20)%5B1d%3A%5D%0A)&g3.tab=0&g3.display_mode=lines&g3.show_exemplars=0&g3.range_input=1h)

The queries below are instant queries

#### Metrics, logs, and traces sent by service name by event count ([link to MLT in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(service_name,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(logs%7Cspans%7Cmetrics)_sent_total%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$1%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22mdai_(logs%7Cspans%7Cmetrics)_sent_total%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

    - sum(increase(mdai_logs_sent_total{}[1h])) by (service_name)
    - sum(increase(mdai_spans_sent_total{}[1h])) by (service_name)
    - sum(increase(mdai_metrics_sent_total{}[1h])) by (service_name)

#### Metrics, logs, and traces received by service name by event count ([link to MLT in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(service_name,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(logs%7Cspans%7Cmetrics)_received_total%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$1%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22mdai_(logs%7Cspans%7Cmetrics)_received_total%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

    - sum(increase(mdai_logs_received_total{}[1h])) by (service_name)
    - sum(increase(mdai_spans_received_total{}[1h])) by (service_name)
    - sum(increase(mdai_metrics_received_total{}[1h])) by (service_name)

#### Metrics, logs, and traces received by service name by packet size ([link to MLT in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(service_name,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_received_total%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$1%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_received_total%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

    - sum(increase(mdai_log_bytes_received_total{}[1h])) by (service_name)
    - sum(increase(mdai_span_bytes_received_total{}[1h])) by (service_name)
    - sum(increase(mdai_metric_bytes_received_total{}[1h])) by (service_name)

#### Metrics, logs, and traces sent by service name by packet size ([link to MLT in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(service_name,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_sent_total%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$1%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22mdai_(log_bytes%7Cspan_bytes%7Cmetric_bytes)_sent_total%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

    - sum(increase(mdai_log_bytes_sent_total{}[1h])) by (service_name)
    - sum(increase(mdai_span_bytes_sent_total{}[1h])) by (service_name)
    - sum(increase(mdai_metric_bytes_sent_total{}[1h])) by (service_name)

## Telemetry tab

### Sankey diagram

The queries below are instant queries

#### Metrics received by receiver ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(receiver,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_receiver_(accepted%7Crefused)_metric_points%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22metrics%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_receiver_(accepted%7Crefused)_metric_points%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_receiver_accepted_metric_points{}[1d])) by (receiver)
- sum(increase(otelcol_receiver_refused_metric_points{}[1d])) by (receiver)

#### Metrics sent by exporter ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(exporter,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_metric_points%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22metrics%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_metric_points%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_exporter_sent_metric_points{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_enqueue_failed_metric_points{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_send_failed_metric_points{}[1d])) by (exporter)

#### Traces received by receiver ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(receiver,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_receiver_(accepted%7Crefused)_spans%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22traces%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_receiver_(accepted%7Crefused)_spans%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_receiver_accepted_spans{}[1d])) by (receiver)
- sum(increase(otelcol_receiver_refused_spans{}[1d])) by (receiver)

#### Traces sent by exporter ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(exporter,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_spans%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22traces%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_spans%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_exporter_sent_spans{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_enqueue_failed_spans{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_send_failed_spans{}[1d])) by (exporter)

#### Logs received by receiver ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(receiver,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_receiver_(accepted%7Crefused)_log_records%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22logs%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_receiver_(accepted%7Crefused)_log_records%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_receiver_accepted_log_records{}[1d])) by (receiver)
- sum(increase(otelcol_receiver_refused_log_records{}[1d])) by (receiver)

#### Logs sent by exporter ([link to see all queries below in one graph](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(exporter,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_log_records%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22log_records%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_log_records%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>))

- sum(increase(otelcol_exporter_enqueue_failed_log_records{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_send_failed_log_records{}[1d])) by (exporter)
- sum(increase(otelcol_exporter_sent_log_records{}[1d])) by (exporter)

To see all data by receiver click [here](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(receiver,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_receiver_(accepted%7Crefused)_(metric_points%7Cspans%7Clog_records)%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$2%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_receiver_(accepted%7Crefused)_(metric_points%7Cspans%7Clog_records)%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>).

To see all data by exporter click [here](<http://localhost:9090/graph?g0.expr=increase(%0A%20%20sum%20by%20(exporter,%20data_type)%20(%0A%20%20%20%20label_replace(%0A%20%20%20%20%20%20%7B__name__%3D~%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_(metric_points%7Cspans%7Clog_records)%22,%20exporter!~%22.*mdai.*%22%7D,%0A%20%20%20%20%20%20%22data_type%22,%0A%20%20%20%20%20%20%22$2%22,%0A%20%20%20%20%20%20%22__name__%22,%0A%20%20%20%20%20%20%22otelcol_exporter_(enqueue_failed%7Csend_failed%7Csent)_(metric_points%7Cspans%7Clog_records)%22%0A%20%20%20%20)%0A%20%20)%5B1d:%5D%0A)&g0.tab=0&g0.display_mode=lines&g0.show_exemplars=0&g0.range_input=1h>).

### Time picker

The queries below are range queries

- sum(increase(otelcol_receiver_accepted_metric_points{}[1d:15m]))
- sum(increase(otelcol_receiver_accepted_log_records{}[1d:15m]))
- sum(increase(otelcol_receiver_accepted_spans{}[1d:15m]))
- sum(increase(otelcol_exporter_sent_metric_point{}[1d:15m]))
- sum(increase(otelcol_exporter_sent_log_record{}[1d:15m]))
- sum(increase(otelcol_exporter_sent_span{}[1d:15m]))
